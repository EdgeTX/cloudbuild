package source

import (
	"context"
	"fmt"
	"os/exec"
	"path"

	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/ldez/go-git-cmd-wrapper/v2/checkout"
	"github.com/ldez/go-git-cmd-wrapper/v2/fetch"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	ginit "github.com/ldez/go-git-cmd-wrapper/v2/init"
	"github.com/ldez/go-git-cmd-wrapper/v2/remote"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
	log "github.com/sirupsen/logrus"
)

type GitDownloader struct {
	GitExecutor types.Executor
	stdOutput   *buildlogs.Recorder
}

func NewGitDownloader(workingDir string, stdOutput *buildlogs.Recorder) *GitDownloader {
	return &GitDownloader{
		GitExecutor: DefaultGitExecutor(workingDir),
		stdOutput:   stdOutput,
	}
}

func DefaultGitExecutor(workingDir string) types.Executor {
	return func(ctx context.Context, name string, debug bool, args ...string) (string, error) {
		log.Tracef("git cmd: %s %s", name, args)
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = workingDir
		if len(args) > 0 && args[0] != "clone" {
			cmd.Env = []string{
				fmt.Sprintf("GIT_WORK_TREE=%s", workingDir),
				fmt.Sprintf("GIT_DIR=%s", path.Join(workingDir, ".git")),
			}
		}
		output, err := cmd.CombinedOutput()
		return string(output), err
	}
}

func (downloader *GitDownloader) init(ctx context.Context) error {
	output, err := git.InitWithContext(
		ctx,
		git.CmdExecutor(downloader.GitExecutor),
		ginit.Directory("."),
	)
	downloader.stdOutput.AddStdOut(output)
	log.Debugf("git init output: %s", output)
	return err
}

func (downloader *GitDownloader) addRemote(ctx context.Context, repository string) error {
	output, err := git.RemoteWithContext(
		ctx,
		git.CmdExecutor(downloader.GitExecutor),
		remote.Add("origin", repository),
	)
	downloader.stdOutput.AddStdOut(output)
	log.Debugf("git init output: %s", output)
	return err
}

func (downloader *GitDownloader) fetch(ctx context.Context, ref string) error {
	output, err := git.FetchWithContext(
		ctx,
		git.CmdExecutor(downloader.GitExecutor),
		fetch.NoTags,
		fetch.Prune,
		fetch.Progress,
		fetch.NoRecurseSubmodules,
		fetch.Depth("1"),
		fetch.Remote("origin"),
		fetch.RefSpec(ref),
	)
	downloader.stdOutput.AddStdOut(output)
	log.Debugf("git fetch output: %s", output)
	return err
}

func (downloader *GitDownloader) checkout(ctx context.Context, ref string) error {
	output, err := git.CheckoutWithContext(
		ctx,
		git.CmdExecutor(downloader.GitExecutor),
		checkout.TreeIsh(ref),
	)
	downloader.stdOutput.AddStdOut(output)
	if err != nil {
		return fmt.Errorf("failed to checkout commit: %w", err)
	}
	log.Debugf("git checkout output: %s", output)
	return nil
}

func (downloader *GitDownloader) updateSubmodules(ctx context.Context) error {
	output, err := git.RawWithContext(ctx, "submodule", func(g *types.Cmd) {
		g.AddOptions("update")
		g.AddOptions("--init")
		g.AddOptions("--recursive")
		g.AddOptions("--depth=1")
	}, git.CmdExecutor(downloader.GitExecutor), git.Debugger(true))
	downloader.stdOutput.AddStdOut(output)
	log.Debugf("git submodules update output: %s", output)
	if err != nil {
		return fmt.Errorf("failed to fetch git submodules: %w", err)
	}
	return nil
}

func (downloader *GitDownloader) Download(ctx context.Context, repository string, commitID string) error {
	err := downloader.init(ctx)
	if err != nil {
		return err
	}

	err = downloader.addRemote(ctx, repository)
	if err != nil {
		return err
	}

	err = downloader.fetch(ctx, commitID)
	if err != nil {
		return err
	}

	err = downloader.checkout(ctx, "FETCH_HEAD")
	if err != nil {
		return err
	}

	err = downloader.updateSubmodules(ctx)
	if err != nil {
		return err
	}

	return nil
}
