package processor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// DiagnosticsInterval is how often disk usage is measured. Walking the
// podman storage is not free, so this is much slower than the heartbeat.
const DiagnosticsInterval = 5 * time.Minute

// diagnosticsCommandTimeout bounds each external command run while
// collecting diagnostics.
const diagnosticsCommandTimeout = time.Minute

// Diagnostics is a snapshot of a worker's resource usage, reported with the
// heartbeat and shown on the admin dashboard. It is meant to find out what
// accumulates on the worker disk over time.
type Diagnostics struct {
	CollectedAt time.Time `json:"collected_at"`
	// StartedAt is when the worker process started.
	StartedAt time.Time `json:"started_at"`
	// JobsProcessed counts jobs run since StartedAt.
	JobsProcessed int64 `json:"jobs_processed"`

	// Filesystem holding the worker container (i.e. the VM disk).
	DiskTotalBytes uint64 `json:"disk_total_bytes"`
	DiskFreeBytes  uint64 `json:"disk_free_bytes"`

	// Space used by known locations inside the worker container.
	TmpBytes           int64 `json:"tmp_bytes"`
	SourceDirs         int   `json:"source_dirs"`
	PodmanStorageBytes int64 `json:"podman_storage_bytes"`

	PodmanImages         int `json:"podman_images"`
	PodmanDanglingImages int `json:"podman_dangling_images"`
	PodmanContainers     int `json:"podman_containers"`
	PodmanVolumes        int `json:"podman_volumes"`

	// Errors lists the measurements that failed, the others remain valid.
	Errors []string `json:"errors,omitempty"`
}

var (
	workerStartedAt = time.Now()
	jobsProcessed   atomic.Int64

	latestDiagnosticsMu sync.Mutex
	latestDiagnostics   []byte
)

func countJobProcessed() {
	jobsProcessed.Add(1)
}

// LatestDiagnostics returns the last collected diagnostics as JSON, or nil
// if none were collected yet.
func LatestDiagnostics() []byte {
	latestDiagnosticsMu.Lock()
	defer latestDiagnosticsMu.Unlock()
	return latestDiagnostics
}

// RunDiagnostics collects diagnostics now and then every DiagnosticsInterval.
func RunDiagnostics(ctx context.Context) {
	for {
		d := CollectDiagnostics(ctx)
		if data, err := json.Marshal(d); err != nil {
			log.Errorf("failed to marshal diagnostics: %s", err)
		} else {
			latestDiagnosticsMu.Lock()
			latestDiagnostics = data
			latestDiagnosticsMu.Unlock()
		}
		if len(d.Errors) > 0 {
			log.Warnf("diagnostics errors: %s", strings.Join(d.Errors, "; "))
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(DiagnosticsInterval):
		}
	}
}

// CollectDiagnostics measures the current resource usage of the worker.
func CollectDiagnostics(ctx context.Context) *Diagnostics {
	d := &Diagnostics{
		CollectedAt:   time.Now(),
		StartedAt:     workerStartedAt,
		JobsProcessed: jobsProcessed.Load(),
	}
	fail := func(what string, err error) {
		d.Errors = append(d.Errors, fmt.Sprintf("%s: %s", what, err))
	}

	var st syscall.Statfs_t
	if err := syscall.Statfs("/", &st); err != nil {
		fail("statfs", err)
	} else {
		d.DiskTotalBytes = st.Blocks * uint64(st.Bsize)
		d.DiskFreeBytes = st.Bavail * uint64(st.Bsize)
	}

	if dirs, err := filepath.Glob(filepath.Join(sourceDirParent, sourceDirPattern)); err != nil {
		fail("source dirs", err)
	} else {
		d.SourceDirs = len(dirs)
	}

	graphRoot, err := podmanOutput(ctx, "info", "--format", "{{.Store.GraphRoot}}")
	if err != nil {
		fail("podman info", err)
	}

	// Run du inside podman's user namespace: files created by containers
	// may be owned by sub-UIDs the worker user cannot read otherwise.
	duPaths := []string{sourceDirParent}
	if graphRoot != "" {
		duPaths = append(duPaths, graphRoot)
	}
	if sizes, err := diskUsage(ctx, duPaths...); err != nil {
		fail("du", err)
	} else {
		d.TmpBytes = sizes[sourceDirParent]
		if graphRoot != "" {
			d.PodmanStorageBytes = sizes[graphRoot]
		}
	}

	counts := []struct {
		name string
		dst  *int
		args []string
	}{
		{"images", &d.PodmanImages, []string{"images", "--quiet", "--no-trunc"}},
		{"dangling images", &d.PodmanDanglingImages, []string{"images", "--quiet", "--no-trunc", "--filter", "dangling=true"}},
		{"containers", &d.PodmanContainers, []string{"ps", "--all", "--quiet"}},
		{"volumes", &d.PodmanVolumes, []string{"volume", "ls", "--quiet"}},
	}
	for _, c := range counts {
		out, err := podmanOutput(ctx, c.args...)
		if err != nil {
			fail("podman "+c.name, err)
			continue
		}
		*c.dst = countLines(out)
	}

	return d
}

func podmanOutput(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, diagnosticsCommandTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "podman", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// diskUsage returns the disk space used by each path, in bytes.
func diskUsage(ctx context.Context, paths ...string) (map[string]int64, error) {
	args := append([]string{"unshare", "du", "--summarize", "--block-size=1"}, paths...)
	out, err := podmanOutput(ctx, args...)
	// du exits with an error when some entries cannot be read (e.g. a file
	// removed while walking), while still printing totals: only fail when
	// nothing could be measured.
	sizes := parseDu(out)
	if len(sizes) == 0 && err != nil {
		return nil, err
	}
	return sizes, nil
}

func parseDu(out string) map[string]int64 {
	sizes := map[string]int64{}
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		size, path, ok := strings.Cut(scanner.Text(), "\t")
		if !ok {
			continue
		}
		if n, err := strconv.ParseInt(size, 10, 64); err == nil {
			sizes[path] = n
		}
	}
	return sizes
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}
