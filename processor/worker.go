package processor

import (
	"context"
	"os"
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

func (worker *Worker) build(
	ctx context.Context,
	job *artifactory.BuildJobModel,
) (*artifactory.BuildJobModel, error) {
	sourceDir, err := os.MkdirTemp("/tmp", "source")
	if err != nil {
		log.Fatalf("failed to create tmp dir: %s", err)
	}
	defer os.RemoveAll(sourceDir)

	recorder := buildlogs.NewRecorder()
	gitDownloader := source.NewGitDownloader(sourceDir, recorder)
	firmwareBuilder := firmware.NewPodmanBuilder(sourceDir, recorder, 2, 1024*1024*1024)

	return worker.artifactory.Build(
		ctx, job, recorder, gitDownloader, firmwareBuilder,
	)
}

func (worker *Worker) executeJob(job *artifactory.BuildJobModel) {
	log.Debugf("starting %s job, result: %s", job.ID, job.Status)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*15)
	defer cancel()

	waitCh := make(chan struct{})
	go func() {
		if _, err := worker.build(ctx, job); err != nil {
			log.Errorf("failed to process next build job: %s", err)
		}
		close(waitCh)
	}()

	select {
	case <-ctx.Done(): // timeout
		log.Errorf("job %s timed out! (status: %s)", job.ID, job.Status)

	case <-waitCh: // finished normally
		log.Infof("processed %s job, result: %s", job.ID, job.Status)
	}
}

func (worker *Worker) PullImage(ctx context.Context, buildImage string) error {
	/*
		We do this so actual build process is faster because of the cached build image
	*/
	recorder := buildlogs.NewRecorder()
	firmwareBuilder := firmware.NewPodmanBuilder("/tmp", recorder, 2, 1024*1024*1024)
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
