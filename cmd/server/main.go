package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edgetx/cloudbuild/artifactory"
	"github.com/edgetx/cloudbuild/server"
	"github.com/edgetx/cloudbuild/storage"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	app := server.New(artifactoryService)
	portRaw := os.Getenv("PORT")
	if len(portRaw) == 0 {
		log.Fatalf("http server port is not defined")
	}
	port, err := strconv.Atoi(portRaw)
	if err != nil {
		log.Fatalf("failed to parse port %s: %s", portRaw, err)
	}
	go func() {
		err := app.Start(port)
		if err != nil {
			log.Fatalf("failed to start http server: %s", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Stop(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %s", err)
	}
}
