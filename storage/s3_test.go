package storage_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edgetx/cloudbuild/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestS3Upload(t *testing.T) {
	s3Mock := &S3MockClient{}
	s3Mock.On("PutObject", mock.Anything, mock.Anything, mock.Anything).
		Return(&s3.PutObjectOutput{}, nil)

	artifactStorage := storage.NewS3ArtifactStorage("test-bucket", s3Mock)
	fileName := "f79982d9968ef7fe4c5c23d9b9e9b200f30e38c28f68601973b98cf702c952e9.bin"

	err := artifactStorage.Upload(context.Background(), []byte("bob"), fileName)
	assert.Nil(t, err)
}
