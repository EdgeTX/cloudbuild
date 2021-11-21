package cli

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/edgetx/cloudbuild/firmware"
	"github.com/edgetx/cloudbuild/source"
	"github.com/pkg/errors"
)

func ParseBuildFlagsFile(buildFlagsFile string) ([]firmware.BuildFlag, error) {
	dataFile, err := os.Open(buildFlagsFile)
	if err != nil {
		return nil, err
	}
	defer dataFile.Close()

	byteValue, err := ioutil.ReadAll(dataFile)
	if err != nil {
		return nil, err
	}

	var result []firmware.BuildFlag
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func Build(
	ctx context.Context,
	buildImage string,
	sourceRepository string,
	commitHash string,
	buildFlagsFile string,
) ([]byte, error) {
	buildFlags, err := ParseBuildFlagsFile(buildFlagsFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse build flags file")
	}

	sourceDir, err := ioutil.TempDir("/tmp", "edgetxsource")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tmp dir")
	}
	defer os.RemoveAll(sourceDir)
	recorder := buildlogs.NewRecorder()
	gitDownloader := source.NewGitDownloader(sourceDir, recorder)
	firmwareBuilder := firmware.NewPodmanBuilder(sourceDir, recorder, runtime.NumCPU(), 1024*1024*1024*2)

	err = gitDownloader.Download(ctx, sourceRepository, commitHash)
	if err != nil {
		return nil, err
	}

	firmwareBin, err := firmwareBuilder.Build(ctx, buildImage, buildFlags)
	if err != nil {
		return nil, err
	}

	return firmwareBin, nil
}
