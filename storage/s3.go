package storage

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edgetx/cloudbuild/config"
	log "github.com/sirupsen/logrus"
)

type PutObjectToS3 interface {
	PutObject(
		ctx context.Context,
		params *s3.PutObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.PutObjectOutput, error)
}

type S3ArtifactStorage struct {
	bucket string
	s3     PutObjectToS3
}

func NewS3ArtifactStorage(bucket string, s3 PutObjectToS3) *S3ArtifactStorage {
	return &S3ArtifactStorage{
		bucket: bucket,
		s3:     s3,
	}
}

func NewS3ArtifactStorageFromConfig(
	ctx context.Context, c *config.CloudbuildOpts,
) *S3ArtifactStorage {
	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               c.StorageS3URL,
				SigningRegion:     region,
				HostnameImmutable: c.StorageS3HostImmutable,
			}, nil
		})

	cfg, err := awsCfg.LoadDefaultConfig(
		ctx,
		awsCfg.WithEndpointResolverWithOptions(resolver),
		awsCfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				c.StorageS3AccessKey, c.StorageS3SecretKey, "",
			),
		),
	)
	if err != nil {
		log.Fatalf("failed to load s3 config: %s", err)
	}
	s3Client := s3.NewFromConfig(cfg)

	return NewS3ArtifactStorage(c.StorageS3Bucket, s3Client)
}

func (storage *S3ArtifactStorage) Upload(ctx context.Context, data []byte, fileName string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(storage.bucket),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(data),
		ACL:    "public-read",
	}

	_, err := storage.s3.PutObject(ctx, input)
	return err
}
