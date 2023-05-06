package source_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/edgetx/cloudbuild/source"
	"github.com/stretchr/testify/assert"
)

func TestFirmwareDownload(t *testing.T) {
	t.Skip()
	sourceDir, err := os.MkdirTemp("/tmp", "source")
	assert.Nil(t, err, "failed to create temp dir for firmware download")
	defer os.RemoveAll(sourceDir)

	repository := "https://github.com/EdgeTX/edgetx.git"
	recorder := buildlogs.NewRecorder()
	firmwareDownloader := source.NewGitDownloader(sourceDir, recorder)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	err = firmwareDownloader.Download(ctx, repository, "8620fe19289c36b574ab68008145a530d589f0fd")
	assert.Nil(t, err, "failed to download repo")
}
