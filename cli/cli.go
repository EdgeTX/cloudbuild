package cli

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path"
	"runtime"

	log "github.com/sirupsen/logrus"

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

	byteValue, err := io.ReadAll(dataFile)
	if err != nil {
		return nil, err
	}

	return ParseBuildFlags(byteValue)
}

func ParseBuildFlags(buildFlagsJSON []byte) ([]firmware.BuildFlag, error) {
	var result []firmware.BuildFlag
	err := json.Unmarshal(buildFlagsJSON, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type BuildCommandArgs struct {
	Target           string
	VersionTag       string
	CommitHash       string
	BuildFlagsFile   string
	BuildFlagsInline string
	BuildImage       string
	SourceRepository string
	LogLevel         log.Level
	ArtifactLocation string
}

func ParseArgs() (*BuildCommandArgs, error) {
	config := &BuildCommandArgs{}

	flag.StringVar(&config.Target, "target", "", "firmware target")
	flag.StringVar(&config.VersionTag, "version", "", "version tag")
	flag.StringVar(&config.CommitHash, "commit", "", "specify commit hash")
	flag.StringVar(&config.BuildFlagsFile, "build-flags-file", "", "specify build flags json file location")
	flag.StringVar(&config.BuildFlagsInline, "build-flags", "", "specify build flags inline")
	flag.StringVar(
		&config.BuildImage,
		"build-image",
		"ghcr.io/edgetx/edgetx-builder",
		"specify podman image for building",
	)

	flag.StringVar(
		&config.SourceRepository,
		"source-repository",
		"https://github.com/EdgeTX/edgetx.git",
		"specify source git repository",
	)

	var logLevelFlag string
	flag.StringVar(
		&logLevelFlag,
		"log-level",
		"debug",
		"specify log level: panic|fatal|error|warn|warning|info|debug|trace",
	)
	logLevel, err := log.ParseLevel(logLevelFlag)
	if err != nil {
		logLevel = log.DebugLevel
	}
	config.LogLevel = logLevel

	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current working directory")
	}
	flag.StringVar(&config.ArtifactLocation, "artifact-location", path.Join(cwd, "./firmware.bin"), "artifact location")

	flag.Parse()

	return config, nil
}

func ValidateBuildArgs(config *BuildCommandArgs) error {
	if len(config.CommitHash) == 0 {
		return errors.New("commit hash is not specified")
	}

	if len(config.BuildFlagsFile) == 0 && len(config.BuildFlagsInline) == 0 {
		return errors.New("you need to specify build flags file location or build flags directly")
	}

	if len(config.BuildFlagsFile) > 0 && len(config.BuildFlagsInline) > 0 {
		return errors.New("can not specify both build flags file and inline build flags params")
	}

	return nil
}

func Build(
	ctx context.Context,
	target string,
	versionTag string,
	buildImage string,
	sourceRepository string,
	commitHash string,
	buildFlags []firmware.BuildFlag,
) ([]byte, error) {
	sourceDir, err := os.MkdirTemp("/tmp", "edgetxsource")
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

	firmwareBin, err := firmwareBuilder.Build(ctx, buildImage, target, versionTag, buildFlags)
	if err != nil {
		return nil, err
	}

	return firmwareBin, nil
}
