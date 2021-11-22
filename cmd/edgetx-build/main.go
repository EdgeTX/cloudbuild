package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/edgetx/cloudbuild/cli"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	var commitHash string
	flag.StringVar(&commitHash, "commit", "", "specify commit hash")

	var buildFlagsFile string
	flag.StringVar(&buildFlagsFile, "build-flags", "", "specify build flags json file location")

	var buildImage string
	flag.StringVar(&buildImage, "build-image", "ghcr.io/edgetx/edgetx-builder:2.5.1", "specify podman image for building")

	var sourceRepository string
	flag.StringVar(
		&sourceRepository,
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
	log.SetLevel(logLevel)

	var artifactLocation string
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current working directory: %s", err)
	}
	flag.StringVar(&artifactLocation, "artifact-location", path.Join(cwd, "./firmware.bin"), "artifact location")

	flag.Parse()

	if len(commitHash) == 0 {
		log.Errorf("commit hash is not specified: %s", commitHash)
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(buildFlagsFile) == 0 {
		log.Errorf("build flags file is not specified: %s", buildFlagsFile)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func(ctx context.Context) {
		buildFlags, err := cli.ParseBuildFlagsFile(buildFlagsFile)
		if err != nil {
			log.Fatalf("failed to parse build flags file: %s", err)
		}

		firmwareBin, err := cli.Build(ctx, buildImage, sourceRepository, commitHash, buildFlags)
		if err != nil {
			log.Fatalf("failed to build firmware: %s", err)
		}

		err = os.WriteFile(artifactLocation, firmwareBin, 0o600)
		if err != nil {
			log.Fatalf("failed to save built firmware: %s", err)
		}

		log.Info("firmware was built successfully")
		os.Exit(0)
	}(ctx)

	<-done
	log.Info("SIGINT|SIGTERM received. Exiting.")
	os.Exit(1)
}
