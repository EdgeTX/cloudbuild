package firmware_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/edgetx/cloudbuild/buildlogs"
	firmware "github.com/edgetx/cloudbuild/firmware"
	"github.com/edgetx/cloudbuild/source"
	"github.com/stretchr/testify/assert"
)

func TestFirmwareFirmwareBuild(t *testing.T) {
	t.Skip()
	sourceDir, err := ioutil.TempDir("/tmp", "source")
	assert.Nil(t, err, "failed to create source dir")
	defer os.RemoveAll(sourceDir)

	t.Log("will start firmware download")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	repository := "https://github.com/EdgeTX/edgetx.git"
	recorder := buildlogs.NewRecorder()
	firmwareDownloader := source.NewGitDownloader(sourceDir, recorder)
	err = firmwareDownloader.Download(ctx, repository, "8620fe19289c36b574ab68008145a530d589f0fd")
	assert.Nil(t, err, "failed to download firmware")
	t.Log("will start firmware build")

	firmwareBuilder := firmware.NewPodmanBuilder(sourceDir, recorder, 2, 1024*1024*1024)
	flags := []firmware.BuildFlag{
		firmware.NewFlag("DISABLE_COMPANION", "YES"),
		firmware.NewFlag("CMAKE_BUILD_TYPE", "Release"),
		firmware.NewFlag("TRACE_SIMPGMSPACE", "NO"),
		firmware.NewFlag("VERBOSE_CMAKELISTS", "YES"),
		firmware.NewFlag("CMAKE_RULE_MESSAGES", "OFF"),
		firmware.NewFlag("PCB", "X10"),
		firmware.NewFlag("PCBREV", "T16"),
		firmware.NewFlag("INTERNAL_MODULE_MULTI", "ON"),
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Minute*20)
	defer cancel()
	firmwareBin, err := firmwareBuilder.Build(ctx, "ghcr.io/edgetx/edgetx-commit-tests:latest", flags)
	assert.Nil(t, err, "failed to build firmware")
	assert.True(t, len(firmwareBin) > 0, "firmware bin is empty")
}
