package firmware

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type PodmanExecutor func(ctx context.Context, args ...string) (string, error)

func DefaultPodmanExecutor(workingDir string) PodmanExecutor {
	return func(ctx context.Context, args ...string) (string, error) {
		log.Debugf("podman cmd: %s", args)
		cmd := exec.CommandContext(ctx, "podman", args...)
		cmd.Dir = workingDir
		output, err := cmd.CombinedOutput()
		return string(output), err
	}
}

type PodmanBuilder struct {
	workingDir     string
	PodmanExecutor PodmanExecutor
	CPULimit       int // logical cores
	recorder       *buildlogs.Recorder
}

func NewPodmanBuilder(workingDir string, recorder *buildlogs.Recorder, cpuLimit int, memoryLimit int) *PodmanBuilder {
	return &PodmanBuilder{
		workingDir:     workingDir,
		PodmanExecutor: DefaultPodmanExecutor(workingDir),
		CPULimit:       cpuLimit,
		recorder:       recorder,
	}
}

func (builder *PodmanBuilder) PullImage(ctx context.Context, buildContainer string) error {
	output, err := builder.PodmanExecutor(ctx, "pull", "--quiet", buildContainer)
	builder.recorder.AddStdOut(output)
	if err != nil {
		return errors.Errorf("failed to pull container image: %s", err)
	}
	log.Debugf("pulled container image: %s", output)
	return nil
}

func (builder *PodmanBuilder) buildCmdArgs(
	buildContainer string,
	target string,
	versionTag string,
	flags []BuildFlag,
) []string {
	args := []string{
		"run",
		"--tty",
		"--rm",
		"--userns=keep-id",
		fmt.Sprintf("--cpus=%d", builder.CPULimit),
		"--volume",
		fmt.Sprintf("%s:/home/rootless/src:Z", builder.workingDir),
	}

	env := []string{
		fmt.Sprintf("FLAVOR=%s", target),
		fmt.Sprintf("EXTRA_OPTIONS=%s", CmakeFlags(flags)),
		fmt.Sprintf("MAX_JOBS=%d", builder.CPULimit),
	}
	if versionTag == "nightly" {
		env = append(env, "EDGETX_VERSION_SUFFIX=nightly")
	} else {
		env = append(
			env,
			"EDGETX_VERSION_SUFFIX=cloudbuild",
			fmt.Sprintf("EDGETX_VERSION_TAG=%s", versionTag),
		)
	}
	for _, varDef := range env {
		args = append(args, "--env", varDef)
	}

	return append(args, buildContainer, "./tools/build-gh.sh")
}

func (builder *PodmanBuilder) matchBuildArtefacts(patterns ...string) ([]string, error) {
	var matches []string

	for _, pattern := range patterns {
		found, err := filepath.Glob(filepath.Join(builder.workingDir, pattern))
		if err != nil {
			return nil, fmt.Errorf("error with pattern %s: %w", pattern, err)
		}
		matches = append(matches, found...)
	}
	return matches, nil
}

func (builder *PodmanBuilder) Build(
	ctx context.Context,
	buildContainer string,
	target string,
	versionTag string,
	flags []BuildFlag,
) ([]byte, error) {
	if err := builder.PullImage(ctx, buildContainer); err != nil {
		return nil, err
	}

	args := builder.buildCmdArgs(buildContainer, target, versionTag, flags)
	output, err := builder.PodmanExecutor(ctx, args...)

	builder.recorder.AddStdOut(output)
	log.Debugf("container build output: %s", output)
	if err != nil {
		return nil, fmt.Errorf("failed to build: %w", err)
	}

	firmwarePaths, err := builder.matchBuildArtefacts("*.uf2", "*.bin")
	if err != nil || len(firmwarePaths) == 0 {
		return nil, fmt.Errorf("cannot find build artifact: %w", err)
	}

	firmwareData, err := os.ReadFile(firmwarePaths[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read firmware binary data: %w", err)
	}

	return firmwareData, nil
}
