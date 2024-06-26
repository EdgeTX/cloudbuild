package artifactory_test

import (
	"context"

	"github.com/edgetx/cloudbuild/firmware"
	"github.com/stretchr/testify/mock"
)

type MockDownloader struct {
	mock.Mock
}

func (downloader *MockDownloader) Download(ctx context.Context, repository string, commitID string) error {
	args := downloader.Called(ctx, repository, commitID)
	return args.Error(0)
}

type MockFirmwareBuilder struct {
	mock.Mock
}

func (downloader *MockFirmwareBuilder) PullImage(ctx context.Context, buildContainer string) error {
	args := downloader.Called(ctx, buildContainer)
	return args.Error(0)
}

func (downloader *MockFirmwareBuilder) Build(
	ctx context.Context,
	buildContainer string,
	target string,
	versionTag string,
	flags []firmware.BuildFlag,
) ([]byte, error) {
	args := downloader.Called(ctx, buildContainer, flags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	out, ok := args.Get(0).([]byte)
	if !ok {
		return nil, args.Error(1)
	}
	return out, args.Error(1)
}

type MockStorage struct {
	mock.Mock
}

func (storage *MockStorage) Upload(ctx context.Context, data []byte, fileName string) error {
	args := storage.Called(ctx, data, fileName)
	return args.Error(0)
}
