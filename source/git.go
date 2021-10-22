package source

import (
	"context"
	"fmt"
	"os/exec"
	"path"

	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/ldez/go-git-cmd-wrapper/v2/checkout"
	"github.com/ldez/go-git-cmd-wrapper/v2/clone"
	"github.com/ldez/go-git-cmd-wrapper/v2/fetch"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/reset"
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

func (downloader *GitDownloader) repositoryExists(ctx context.Context) bool {
	output, err := git.StatusWithContext(ctx, git.CmdExecutor(downloader.GitExecutor))
	downloader.stdOutput.AddStdOut(output)
	log.Debugf("git status output: %s", output)
	return err == nil
}

func (downloader *GitDownloader) cloneRepository(ctx context.Context, repository string) error {
	output, err := git.CloneWithContext(
		ctx,
		git.CmdExecutor(downloader.GitExecutor),
		clone.Repository(repository),
		clone.NoCheckout,
		clone.ShallowSubmodules,
		clone.NoHardlinks,
		func(g *types.Cmd) {
			g.AddOptions("--filter=blob:none")
		},
		clone.Directory("."),
	)
	downloader.stdOutput.AddStdOut(output)
	log.Debugf("git clone output: %s", output)
	return err
}

func (downloader *GitDownloader) fetchRepository(ctx context.Context) error {
	output, err := git.FetchWithContext(ctx, git.CmdExecutor(downloader.GitExecutor), fetch.Tags)
	downloader.stdOutput.AddStdOut(output)
	log.Debugf("git fetch output: %s", output)
	return fmt.Errorf("failed to fetch repository: %w", err)
}

func (downloader *GitDownloader) resetRepository(ctx context.Context) error {
	output, err := git.ResetWithContext(ctx, git.CmdExecutor(downloader.GitExecutor), reset.Hard)
	downloader.stdOutput.AddStdOut(output)
	log.Debugf("git reset output: %s", output)
	return fmt.Errorf("failed to reset repository: %w", err)
}

func (downloader *GitDownloader) setupSparseCheckout(ctx context.Context) error {
	output, err := git.RawWithContext(ctx, "sparse-checkout", func(g *types.Cmd) {
		g.AddOptions("init")
		g.AddOptions("--cone")
	}, git.CmdExecutor(downloader.GitExecutor))
	downloader.stdOutput.AddStdOut(output)
	log.Debugf("git sparse checkout init output: %s", output)
	if err != nil {
		return fmt.Errorf("failed to init sparse checkout: %w", err)
	}

	output, err = git.RawWithContext(ctx, "sparse-checkout", func(g *types.Cmd) {
		g.AddOptions("set")
		g.AddOptions("/radio")
		g.AddOptions("/tools")
		g.AddOptions("/cmake")
	}, git.CmdExecutor(downloader.GitExecutor), git.Debugger(true))
	downloader.stdOutput.AddStdOut(output)
	log.Debugf("git sparse checkout output: %s", output)
	if err != nil {
		return fmt.Errorf("failed to set sparse-checkout: %w", err)
	}
	return nil
}

func (downloader *GitDownloader) checkout(ctx context.Context, commitID string) error {
	output, err := git.CheckoutWithContext(ctx, git.CmdExecutor(downloader.GitExecutor), checkout.Branch(commitID))
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
	if downloader.repositoryExists(ctx) {
		err := downloader.resetRepository(ctx)
		if err != nil {
			return err
		}

		err = downloader.fetchRepository(ctx)
		if err != nil {
			return err
		}
	} else {
		err := downloader.cloneRepository(ctx, repository)
		if err != nil {
			return err
		}

		err = downloader.setupSparseCheckout(ctx)
		if err != nil {
			return err
		}

		err = downloader.checkout(ctx, commitID)
		if err != nil {
			return err
		}
	}

	err := downloader.updateSubmodules(ctx)
	if err != nil {
		return err
	}

	return nil
}
