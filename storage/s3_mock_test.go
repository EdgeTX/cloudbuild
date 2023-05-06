package storage_test

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/mock"
)

type S3MockClient struct {
	mock.Mock
}

func (s3MockClient *S3MockClient) PutObject(
	ctx context.Context,
	params *s3.PutObjectInput,
	optFns ...func(*s3.Options),
) (*s3.PutObjectOutput, error) {
	args := s3MockClient.Called(ctx, params, optFns)
	out, ok := args.Get(0).(*s3.PutObjectOutput)
	if !ok {
		return nil, args.Error(1)
	}
	return out, args.Error(1)
}
