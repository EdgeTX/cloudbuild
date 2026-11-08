package firmware

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type PodmanExecutor func(ctx context.Context, args ...string) (string, error)

func DefaultPodmanExecutor(workingDir string) PodmanExecutor {
	return func(ctx context.Context, args ...string) (string, error) {
		log.Debugf("podman cmd: %s", args)
		cmd := exec.CommandContext(ctx, "podman", args...)
		cmd.Dir = workingDir
		// Don't block forever on pipes held open by processes podman
		// may leave behind (e.g. conmon) once the command is killed.
		cmd.WaitDelay = 10 * time.Second
		output, err := cmd.CombinedOutput()
		return string(output), err
	}
}

type PodmanBuilder struct {
	workingDir     string
	PodmanExecutor PodmanExecutor
	CPULimit       int // logical cores
	recorder       *buildlogs.Recorder
}

func NewPodmanBuilder(workingDir string, recorder *buildlogs.Recorder, cpuLimit int, memoryLimit int) *PodmanBuilder {
	return &PodmanBuilder{
		workingDir:     workingDir,
		PodmanExecutor: DefaultPodmanExecutor(workingDir),
		CPULimit:       cpuLimit,
		recorder:       recorder,
	}
}

func (builder *PodmanBuilder) PullImage(ctx context.Context, buildContainer string) error {
	output, err := builder.PodmanExecutor(ctx, "pull", "--quiet", buildContainer)
	builder.recorder.AddStdOut(output)
	if err != nil {
		return errors.Errorf("failed to pull container image: %s", err)
	}
	log.Debugf("pulled container image: %s", output)
	builder.pruneDanglingImages(ctx)
	return nil
}

// pruneDanglingImages removes untagged images left behind when a floating
// tag (e.g. "edgetx-builder:2.12" or ":latest") is re-pulled after being
// updated upstream. Without this, every upstream update leaves a multi-GB
// image in the worker's podman storage until the disk fills up.
// Images still used by a container are kept by podman.
func (builder *PodmanBuilder) pruneDanglingImages(ctx context.Context) {
	output, err := builder.PodmanExecutor(ctx, "image", "prune", "--force")
	if err != nil {
		log.Warnf("failed to prune dangling images: %s: %s", err, output)
		return
	}
	if output != "" {
		log.Infof("pruned dangling images: %s", output)
	}
}

// containerCleanupTimeout bounds the time spent force-removing a build
// container. It is independent of the build context, which is usually
// already expired when the cleanup runs.
const containerCleanupTimeout = time.Minute

// containerName returns a name unique to this build, derived from its
// (unique) working directory, so the container can be found and removed
// even if the podman client was killed.
func (builder *PodmanBuilder) containerName() string {
	return "cloudbuild-" + filepath.Base(builder.workingDir)
}

// removeContainer stops (if still running) and removes the build container
// and its anonymous volumes. It returns only once podman is done, so nothing
// writes into the working directory afterwards.
func (builder *PodmanBuilder) removeContainer() {
	ctx, cancel := context.WithTimeout(context.Background(), containerCleanupTimeout)
	defer cancel()
	output, err := builder.PodmanExecutor(ctx,
		"rm", "--force", "--ignore", "--volumes", "--time", "0",
		builder.containerName(),
	)
	if err != nil {
		log.Errorf("failed to remove build container %s: %s: %s",
			builder.containerName(), err, output)
	}
}

func (builder *PodmanBuilder) buildCmdArgs(
	ctx context.Context,
	buildContainer string,
	target string,
	versionTag string,
	flags []BuildFlag,
) []string {
	args := []string{
		"run",
		"--tty",
		"--rm",
		"--userns=keep-id",
		"--name", builder.containerName(),
		fmt.Sprintf("--cpus=%d", builder.CPULimit),
		"--volume",
		fmt.Sprintf("%s:/home/rootless/src:Z", builder.workingDir),
	}

	// Let conmon kill the container itself when the build deadline is
	// reached, independently of the podman client being alive.
	if deadline, ok := ctx.Deadline(); ok {
		if secs := int(time.Until(deadline).Seconds()); secs > 0 {
			args = append(args, fmt.Sprintf("--timeout=%d", secs))
		}
	}

	env := []string{
		fmt.Sprintf("FLAVOR=%s", target),
		fmt.Sprintf("EXTRA_OPTIONS=%s", CmakeFlags(flags)),
		fmt.Sprintf("MAX_JOBS=%d", builder.CPULimit),
		fmt.Sprintf("CMAKE_BUILD_PARALLEL_LEVEL=%d", builder.CPULimit),
	}
	if versionTag == "nightly" {
		env = append(env, "EDGETX_VERSION_SUFFIX=nightly")
	} else {
		env = append(
			env,
			"EDGETX_VERSION_SUFFIX=cloudbuild",
			fmt.Sprintf("EDGETX_VERSION_TAG=%s", versionTag),
		)
	}
	for _, varDef := range env {
		args = append(args, "--env", varDef)
	}

	return append(args, buildContainer, "./tools/build-gh.sh")
}

func (builder *PodmanBuilder) matchBuildArtefacts(patterns ...string) ([]string, error) {
	var matches []string

	for _, pattern := range patterns {
		found, err := filepath.Glob(filepath.Join(builder.workingDir, pattern))
		if err != nil {
			return nil, fmt.Errorf("error with pattern %s: %w", pattern, err)
		}
		matches = append(matches, found...)
	}
	return matches, nil
}

func (builder *PodmanBuilder) Build(
	ctx context.Context,
	buildContainer string,
	target string,
	versionTag string,
	flags []BuildFlag,
) ([]byte, error) {
	if err := builder.PullImage(ctx, buildContainer); err != nil {
		return nil, err
	}

	args := builder.buildCmdArgs(ctx, buildContainer, target, versionTag, flags)
	output, err := builder.PodmanExecutor(ctx, args...)
	// With --rm this is a no-op on normal exit. When the podman client was
	// killed (build timeout), the container would otherwise keep running,
	// writing into the working directory while it is being deleted.
	builder.removeContainer()

	builder.recorder.AddStdOut(output)
	log.Debugf("container build output: %s", output)
	if err != nil {
		return nil, fmt.Errorf("failed to build: %w", err)
	}

	firmwarePaths, err := builder.matchBuildArtefacts("*.uf2", "*.bin")
	if err != nil || len(firmwarePaths) == 0 {
		return nil, fmt.Errorf("cannot find build artifact: %w", err)
	}

	firmwareData, err := os.ReadFile(firmwarePaths[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read firmware binary data: %w", err)
	}

	return firmwareData, nil
}

// CleanupStaleContainers removes every podman container and unused volume
// left over by previous worker runs (e.g. after a crash or a timed-out
// build). It must only be called when no build is in progress: the worker
// owns its podman storage and does not run anything else in it.
func CleanupStaleContainers(ctx context.Context) {
	podman := DefaultPodmanExecutor("/tmp")
	if output, err := podman(ctx, "rm", "--all", "--force", "--volumes", "--time", "0"); err != nil {
		log.Warnf("failed to remove stale containers: %s: %s", err, output)
	} else if output != "" {
		log.Infof("removed stale containers: %s", output)
	}
	if output, err := podman(ctx, "volume", "prune", "--force"); err != nil {
		log.Warnf("failed to prune volumes: %s: %s", err, output)
	} else if output != "" {
		log.Infof("pruned volumes: %s", output)
	}
}
