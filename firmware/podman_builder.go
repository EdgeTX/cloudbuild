package firmware

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type PodmanExecutor func(ctx context.Context, name string, args ...string) (string, error)

func DefaultPodmanExecutor(workingDir string) PodmanExecutor {
	return func(ctx context.Context, name string, args ...string) (string, error) {
		log.Debugf("podman cmd: %s %s", name, args)
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = workingDir
		output, err := cmd.CombinedOutput()
		return string(output), err
	}
}

type PodmanBuilder struct {
	workingDir     string
	PodmanExecutor PodmanExecutor
	CPULimit       int // logical cores
	MemoryLimit    int // bytes
	recorder       *buildlogs.Recorder
}

func NewPodmanBuilder(workingDir string, recorder *buildlogs.Recorder, cpuLimit int, memoryLimit int) *PodmanBuilder {
	return &PodmanBuilder{
		workingDir:     workingDir,
		PodmanExecutor: DefaultPodmanExecutor(workingDir),
		CPULimit:       cpuLimit,
		MemoryLimit:    memoryLimit,
		recorder:       recorder,
	}
}

func (builder *PodmanBuilder) PullImage(ctx context.Context, buildContainer string) error {
	output, err := builder.PodmanExecutor(ctx, "podman", "pull", "--quiet", buildContainer)
	builder.recorder.AddStdOut(output)
	if err != nil {
		return errors.Errorf("failed to pull container image: %s", err)
	}
	log.Debugf("pulled container image: %s", output)
	return nil
}

func (builder *PodmanBuilder) Build(ctx context.Context, buildContainer string, flags []BuildFlag) ([]byte, error) {
	err := builder.PullImage(ctx, buildContainer)
	if err != nil {
		return nil, err
	}

	commands := []string{
		"rm -rf ./build",
		"mkdir ./build",
		"cd ./build",
		fmt.Sprintf("cmake -Wno-dev %s ../", CmakeFlags(flags)),
		"cd ../",
		fmt.Sprintf("make --directory ./build -j%d firmware-size", builder.CPULimit*2),
	}
	output, err := builder.PodmanExecutor(
		ctx,
		"podman",
		"run",
		"--tty",
		"--userns=keep-id",
		"--rm",
		fmt.Sprintf("--cpus=%d", builder.CPULimit),
		"--workdir",
		"/home/rootless/src",
		"--volume",
		fmt.Sprintf("%s:/home/rootless/src:Z", builder.workingDir),
		buildContainer,
		"bash",
		"-c",
		strings.Join(commands, " && "),
	)
	builder.recorder.AddStdOut(output)
	log.Debugf("container build output: %s", output)
	if err != nil {
		return nil, fmt.Errorf("failed to build: %w", err)
	}

	firmwarePath := path.Join(builder.workingDir, "build", "arm-none-eabi", "firmware.bin")
	if _, err := os.Stat(firmwarePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("firmware.bin does not exist: %w", err)
	}

	firmwareData, err := os.ReadFile(firmwarePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read firmware binary data: %w", err)
	}

	return firmwareData, nil
}
