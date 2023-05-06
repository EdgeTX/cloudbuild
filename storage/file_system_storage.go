package storage

import (
	"context"
	"os"
	"path"
)

type FileSystemStorage struct {
	storageFolder string
}

func NewLocalStorage(storageFolder string) *FileSystemStorage {
	return &FileSystemStorage{
		storageFolder: storageFolder,
	}
}

func (storage *FileSystemStorage) Upload(ctx context.Context, data []byte, fileName string) error {
	return os.WriteFile(path.Join(storage.storageFolder, fileName), data, 0o600)
}
