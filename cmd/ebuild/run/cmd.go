package run

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/edgetx/cloudbuild/artifactory"
	"github.com/edgetx/cloudbuild/auth"
	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/edgetx/cloudbuild/config"
	"github.com/edgetx/cloudbuild/firmware"
	"github.com/edgetx/cloudbuild/processor"
	"github.com/edgetx/cloudbuild/server"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type serverRunner struct {
	opts *config.CloudbuildOpts
	ctx  context.Context
}

func (s *serverRunner) initLogging() {
	log.SetOutput(os.Stdout)
	if s.opts.LogLevel != config.DebugLevel {
		gin.SetMode(gin.ReleaseMode)
	}
	log.SetLevel(s.opts.LogLevel.Level())
	log.Debugln("Logging initialised")
}

func (s *serverRunner) runAPI(cmd *cobra.Command, args []string) {
	s.initLogging()
	art, err := artifactory.NewFromConfig(s.ctx, s.opts)
	if err != nil {
		fmt.Printf("failed to create artifactory: %s", err)
		os.Exit(1)
	}
	auth, err := auth.NewAuthTokenDBFromConfig(s.opts)
	if err != nil {
		fmt.Printf("failed to create authenticator: %s", err)
		os.Exit(1)
	}
	app := server.New(art, auth)
	err = app.Start(
		fmt.Sprintf("%s:%d",
			s.opts.HTTPBindAddress,
			s.opts.HTTPBindPort,
		),
	)
	if err != nil {
		fmt.Printf("failed to start API server: %s", err)
		os.Exit(1)
	}
}

func (s *serverRunner) runWorker(cmd *cobra.Command, args []string) {
	s.initLogging()
	art, err := artifactory.NewFromConfig(s.ctx, s.opts)
	if err != nil {
		fmt.Printf("failed to create artifactory: %s", err)
		os.Exit(1)
	}

	worker := processor.New(art)
	/*
		We do this so actual build process is faster because of the cached build image
	*/
	recorder := buildlogs.NewRecorder()
	firmwareBuilder := firmware.NewPodmanBuilder("/tmp", recorder, 2, 1024*1024*1024)
	ctx, cancel := context.WithTimeout(s.ctx, artifactory.MaxBuildDuration)
	defer cancel()
	log.Infof("Will pull current build image for cache")
	err = firmwareBuilder.PullImage(ctx, s.opts.BuildImage)
	if err != nil {
		log.Errorf("pull image logs: %s", recorder.Logs())
		log.Error("failed to pre-pull edgetx build image")
		os.Exit(1)
	} else {
		log.Infof("Image downloaded successfully")
	}

	go func() {
		worker.Run()
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 4 minutes.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(s.ctx, 4*time.Minute)
	defer cancel()

	if err := worker.Stop(ctx); err != nil {
		log.Fatalf("Worker forced to shutdown: %s", err)
	}
}

type cobraRunFunc func(*cobra.Command, []string)

func (s *serverRunner) makeCmd(use, short string, f cobraRunFunc) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run:   f,
	}
}

func newServerRunner(ctx context.Context, o *config.CloudbuildOpts) *serverRunner {
	return &serverRunner{
		opts: o,
		ctx:  ctx,
	}
}

func NewAPICommand(s *serverRunner) *cobra.Command {
	cmd := s.makeCmd("api", "Run the API server", s.runAPI)
	s.opts.BindAPIOpts(cmd)
	return cmd
}

func NewWorkerCommand(s *serverRunner) *cobra.Command {
	cmd := s.makeCmd("worker", "Run a cloudbuild worker", s.runWorker)
	s.opts.BindWorkerOpts(cmd)
	return cmd
}

func NewRunCommand(ctx context.Context, o *config.CloudbuildOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the EdgeTX build server",
	}
	o.BindCliOpts(cmd)
	o.BindDBOpts(cmd)
	o.BindStorageOpts(cmd)

	s := newServerRunner(ctx, o)
	cmd.AddCommand(NewAPICommand(s))
	cmd.AddCommand(NewWorkerCommand(s))

	return cmd
}
