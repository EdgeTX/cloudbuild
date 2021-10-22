package processor

import (
	"context"
	"io/ioutil"
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

func (worker *Worker) Run() {
	worker.running = true
	for worker.running {
		worker.inProgress = true
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*15)
		sourceDir, err := ioutil.TempDir("/tmp", "source")
		if err != nil {
			log.Fatalf("failed to create tmp dir: %s", err)
		}
		defer os.RemoveAll(sourceDir)
		recorder := buildlogs.NewRecorder()
		gitDownloader := source.NewGitDownloader(sourceDir, recorder)
		firmwareBuilder := firmware.NewPodmanBuilder(sourceDir, recorder, 2, 1024*1024*1024)
		job, err := worker.artifactory.ProcessNextBuildJob(ctx, recorder, gitDownloader, firmwareBuilder)
		if err != nil {
			log.Errorf("failed to process next build job: %s", err)
		}
		if job != nil {
			log.Infof("processed %s job, result: %s", job.ID, job.Status)
		}

		cancel()
		time.Sleep(time.Second * 2)
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
