package storage_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/edgetx/cloudbuild/storage"
	"github.com/stretchr/testify/assert"
)

func TestFileSystemStorage(t *testing.T) {
	storageFolder := "/tmp"
	artifactStorage := storage.NewLocalStorage(storageFolder)

	fileName := "f79982d9968ef7fe4c5c23d9b9e9b200f30e38c28f68601973b98cf702c952e9.bin"
	data := []byte("bob")
	err := artifactStorage.Upload(context.Background(), data, fileName)
	assert.Nil(t, err)

	fileInfo, err := os.Stat(path.Join(storageFolder, fileName))
	assert.Nil(t, err)
	assert.Equal(t, int64(len(data)), fileInfo.Size())
}
