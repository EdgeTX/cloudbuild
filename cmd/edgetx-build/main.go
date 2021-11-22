package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgetx/cloudbuild/firmware"

	"github.com/edgetx/cloudbuild/cli"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	config, err := cli.ParseArgs()
	if err != nil {
		log.Fatalf("failed to parse cli flags: %s", err)
	}

	log.SetLevel(config.LogLevel)

	err = cli.ValidateBuildArgs(config)
	if err != nil {
		log.Fatalf("cli flags invalid: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context) {
		var buildFlags []firmware.BuildFlag
		if len(config.BuildFlagsFile) > 0 {
			buildFlags, err = cli.ParseBuildFlagsFile(config.BuildFlagsFile)
			if err != nil {
				log.Fatalf("failed to parse build flags file: %s", err)
			}
		}

		if len(config.BuildFlagsInline) > 0 {
			buildFlags, err = cli.ParseCmakeString(config.BuildFlagsInline)
			if err != nil {
				log.Fatalf("failed to parse build flags file: %s", err)
			}
		}

		firmwareBin, err := cli.Build(ctx, config.BuildImage, config.SourceRepository, config.CommitHash, buildFlags)
		if err != nil {
			log.Fatalf("failed to build firmware: %s", err)
		}

		err = os.WriteFile(config.ArtifactLocation, firmwareBin, 0o600)
		if err != nil {
			log.Fatalf("failed to save built firmware: %s", err)
		}

		log.Info("firmware was built successfully")
		os.Exit(0)
	}(ctx)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	log.Info("SIGINT|SIGTERM received. Exiting.")
	os.Exit(1)
}
