package storage

import (
	"context"

	"github.com/edgetx/cloudbuild/config"
)

const (
	StorageTypeS3  = "S3"
	StorageTypeLFS = "FILE_SYSTEM_STORAGE"
)

type Handler interface {
	Upload(context context.Context, data []byte, fileName string) error
}

func NewFromConfig(ctx context.Context, c *config.CloudbuildOpts) Handler {
	switch c.StorageType {
	case StorageTypeS3:
		return NewS3ArtifactStorageFromConfig(ctx, c)
	case StorageTypeLFS:
		return NewLocalStorage(c.StoragePath)
	}
	return nil
}
