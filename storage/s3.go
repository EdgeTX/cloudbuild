package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3ArtifactStorage struct {
	bucket string
	s3     *s3.Client
}

func NewS3ArtifactStorage(bucket string, s3 *s3.Client) *S3ArtifactStorage {
	return &S3ArtifactStorage{
		bucket: bucket,
		s3:     s3,
	}
}

func (storage *S3ArtifactStorage) Upload(ctx context.Context, data []byte, fileName string) (*url.URL, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(storage.bucket),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(data),
		ACL:    "public-read",
	}

	_, err := storage.s3.PutObject(ctx, input)
	if err != nil {
		return nil, err
	}

	return url.Parse(fmt.Sprintf("https://%s.s3.amazonaws.com/%s", storage.bucket, fileName))
}
