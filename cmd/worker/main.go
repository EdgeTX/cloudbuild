package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edgetx/cloudbuild/artifactory"
	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/edgetx/cloudbuild/firmware"
	"github.com/edgetx/cloudbuild/processor"
	"github.com/edgetx/cloudbuild/storage"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	log.SetOutput(os.Stdout)

	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.Errorf("failed to parse LOG_LEVEL: %s", err)
		logLevel = log.DebugLevel
	}
	log.SetLevel(logLevel)
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load s3 config: %s", err)
	}
	s3Client := s3.NewFromConfig(cfg)
	bucket := os.Getenv("ARTIFACT_STORAGE_S3_BUCKET")
	artifactStorage := storage.NewS3ArtifactStorage(bucket, s3Client)

	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connecto to the database: %s", err)
	}
	buildJobsRepository := artifactory.NewBuildJobsDBRepository(db)

	buildImage := os.Getenv("BUILD_IMAGE")
	if len(buildImage) == 0 {
		log.Fatal("BUILD_IMAG env variable is not specified")
	}

	sourceRepository := os.Getenv("SOURCE_REPOSITORY")
	if len(sourceRepository) == 0 {
		log.Fatal("sourceRepository env variable is not specified")
	}
	artifactoryService := artifactory.New(buildJobsRepository, artifactStorage, buildImage, sourceRepository)

	worker := processor.New(artifactoryService)

	go func() {
		/*
			We do this so actual build process is faster because of the cached build image
		*/
		recorder := buildlogs.NewRecorder()
		firmwareBuilder := firmware.NewPodmanBuilder("/tmp", recorder, 2, 1024*1024*1024)
		ctx, cancel := context.WithTimeout(context.Background(), artifactory.MaxBuildDuration)
		defer cancel()
		log.Infof("Will pull current build image for cache")
		err := firmwareBuilder.PullImage(ctx, buildImage)
		if err != nil {
			log.Errorf("pull image logs: %s", recorder.Logs())
			log.Error("failed to pre-pull edgetx build image")
		}
		log.Infof("Image downloaded successfully")
	}()

	go func() {
		worker.Run()
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 4 minutes.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 4 * time.Minute)
	defer cancel()

	if err := worker.Stop(ctx); err != nil {
		log.Fatalf("Worker forced to shutdown: %s", err)
	}
}
