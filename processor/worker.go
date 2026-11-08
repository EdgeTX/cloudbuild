package processor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/edgetx/cloudbuild/artifactory"
	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/edgetx/cloudbuild/firmware"
	"github.com/edgetx/cloudbuild/source"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Worker struct {
	artifactory *artifactory.Artifactory
	running     bool
	inProgress  bool
}

func New(artifactory *artifactory.Artifactory) *Worker {
	return &Worker{
		artifactory: artifactory,
		running:     false,
		inProgress:  false,
	}
}

const (
	sourceDirParent  = "/tmp"
	sourceDirPattern = "source*"

	// jobCleanupGracePeriod is how long to wait, after a job timed out,
	// for the build to stop and clean up before moving on.
	jobCleanupGracePeriod = 2 * time.Minute
)

// removeSourceDir deletes a build working directory. If a plain removal
// fails (e.g. files owned by a sub-UID created inside the build container),
// it retries from within podman's user namespace.
func removeSourceDir(dir string) {
	err := os.RemoveAll(dir)
	if err == nil {
		return
	}
	log.Warnf("failed to remove source dir %s: %s, retrying with podman unshare", dir, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	output, err := exec.CommandContext(ctx, "podman", "unshare", "rm", "-rf", dir).CombinedOutput()
	if err != nil {
		log.Errorf("failed to remove source dir %s: %s: %s", dir, err, output)
	}
}

// CleanupStaleBuilds removes everything previous worker runs may have left
// behind: build containers, their volumes and source directories.
// It must be called before the worker starts processing jobs.
func CleanupStaleBuilds(ctx context.Context) {
	firmware.CleanupStaleContainers(ctx)

	dirs, err := filepath.Glob(filepath.Join(sourceDirParent, sourceDirPattern))
	if err != nil {
		log.Errorf("failed to list stale source dirs: %s", err)
		return
	}
	for _, dir := range dirs {
		log.Infof("removing stale source dir %s", dir)
		removeSourceDir(dir)
	}
}

func (worker *Worker) build(
	ctx context.Context,
	job *artifactory.BuildJobModel,
) (*artifactory.BuildJobModel, error) {
	sourceDir, err := os.MkdirTemp(sourceDirParent, sourceDirPattern)
	if err != nil {
		log.Fatalf("failed to create tmp dir: %s", err)
	}
	defer removeSourceDir(sourceDir)

	recorder := buildlogs.NewRecorder()
	gitDownloader := source.NewGitDownloader(sourceDir, recorder)
	firmwareBuilder := firmware.NewPodmanBuilder(sourceDir, recorder, runtime.NumCPU(), 2*1024*1024*1024)

	return worker.artifactory.Build(
		ctx, job, recorder, gitDownloader, firmwareBuilder,
	)
}

func (worker *Worker) executeJob(job *artifactory.BuildJobModel) {
	log.Debugf("starting %s job, result: %s", job.ID, job.Status)
	ctx, cancel := context.WithTimeout(context.Background(), artifactory.MaxBuildDuration)
	defer cancel()

	waitCh := make(chan struct{})
	go func() {
		if _, err := worker.build(ctx, job); err != nil {
			log.Errorf("failed to process next build job: %s", err)
		}
		close(waitCh)
	}()

	select {
	case <-waitCh: // finished normally
		log.Infof("processed %s job, result: %s", job.ID, job.Status)
		return

	case <-ctx.Done(): // timeout
		log.Errorf("job %s timed out!", job.ID)
	}

	// Wait for the build to stop its container and remove its source dir,
	// so that it neither leaks disk space nor competes with the next job.
	select {
	case <-waitCh:
		log.Infof("cleaned up timed out job %s", job.ID)
	case <-time.After(jobCleanupGracePeriod):
		log.Errorf("job %s still running %s after timeout, giving up", job.ID, jobCleanupGracePeriod)
	}
}

func (worker *Worker) PullImage(ctx context.Context, buildImage string) error {
	/*
		We do this so actual build process is faster because of the cached build image
	*/
	recorder := buildlogs.NewRecorder()
	firmwareBuilder := firmware.NewPodmanBuilder("/tmp", recorder, runtime.NumCPU(), 1024*1024*1024)
	ctx, cancel := context.WithTimeout(ctx, artifactory.MaxBuildDuration)
	defer cancel()

	log.Infof("Will pull current build image for cache")
	err := firmwareBuilder.PullImage(ctx, buildImage)
	if err != nil {
		log.Errorf("pull image logs: %s", recorder.Logs())
	}
	return err
}

func (worker *Worker) Run() {
	worker.running = true
	for worker.running {
		worker.inProgress = true
		job, err := worker.artifactory.ReservePendingBuild()
		if err != nil {
			log.Errorf("failed to reserve next build job: %s", err)
			time.Sleep(time.Second * 1)
			continue
		}

		if job != nil {
			worker.executeJob(job)
		} else {
			time.Sleep(time.Second * 1)
		}

		worker.inProgress = false
	}
}

func (worker *Worker) Stop(ctx context.Context) error {
	worker.running = false

	shutdownDone := make(chan bool)
	go func() {
		for worker.inProgress {
			time.Sleep(time.Second * 1)
			log.Info("Waiting for processor shutdown...")
		}
		shutdownDone <- true
	}()

	select {
	case <-ctx.Done():
		return errors.New("failed to shutdown worker in time")
	case <-shutdownDone:
		return nil
	}
}
