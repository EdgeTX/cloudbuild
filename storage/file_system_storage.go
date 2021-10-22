package storage

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
)

type FileSystemStorage struct {
	storageFolder string
	downloadURL   string
}

func NewLocalStorage(storageFolder string, downloadURL string) *FileSystemStorage {
	return &FileSystemStorage{
		storageFolder: storageFolder,
		downloadURL:   downloadURL,
	}
}

func (storage *FileSystemStorage) Upload(ctx context.Context, data []byte, fileName string) (*url.URL, error) {
	err := os.WriteFile(path.Join(storage.storageFolder, fileName), data, 0o600)
	if err != nil {
		return nil, err
	}
	return url.Parse(fmt.Sprintf("%s/%s", storage.downloadURL, fileName))
}
