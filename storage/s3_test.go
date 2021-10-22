package storage_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edgetx/cloudbuild/storage"
	"github.com/stretchr/testify/assert"
)

func TestS3Upload(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	assert.Nil(t, err, "failed to load aws config")

	s3Client := s3.NewFromConfig(cfg)

	bucket := os.Getenv("ARTIFACT_STORAGE_S3_BUCKET")
	artifactStorage := storage.NewS3ArtifactStorage(bucket, s3Client)

	fileName := "f79982d9968ef7fe4c5c23d9b9e9b200f30e38c28f68601973b98cf702c952e9.bin"
	url, err := artifactStorage.Upload(context.Background(), []byte("bob"), fileName)
	assert.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, fileName), url.String())
}
