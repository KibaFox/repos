package git

import (
	"bytes"
	"context"
	"os/exec"
	"strings"

	"gitlab.com/kibafox/repos/internal/errs"
)

const gitCmd = "git"

// Run will run git with the provided arguments.
func Run(ctx context.Context, args ...string) error {
	bufErr := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, gitCmd, args...)
	cmd.Stderr = bufErr

	if err := cmd.Run(); err != nil {
		return errs.NewErrGit(bufErr, args...)
	}

	return nil
}

// Out will run git with the provided arguments and return the captured output.
func Out(ctx context.Context, args ...string) (string, error) {
	bufOut := &bytes.Buffer{}
	bufErr := &bytes.Buffer{}

	cmd := exec.CommandContext(ctx, gitCmd, args...)
	cmd.Stdout = bufOut
	cmd.Stderr = bufErr

	if err := cmd.Run(); err != nil {
		return "", errs.NewErrGit(bufErr, args...)
	}

	return strings.TrimSuffix(bufOut.String(), "\n"), nil
}

func bol(ctx context.Context, args ...string) bool {
	cmd := exec.CommandContext(ctx, gitCmd, args...)
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func Clone(ctx context.Context, remote, local string) error {
	return Run(ctx, "clone", "--quiet", remote, local)
}

func Pull(ctx context.Context, path string) error {
	return Run(ctx, "-C", path, "pull", "--ff-only", "--quiet")
}

func Origin(ctx context.Context, path string) (string, error) {
	return Out(ctx, "-C", path, "remote", "get-url", "origin")
}

func Dirty(ctx context.Context, path string) bool {
	return !bol(ctx, "-C", path, "diff",
		"--no-ext-diff", "--quiet", "--exit-code")
}

func Staged(ctx context.Context, path string) bool {
	return !bol(ctx, "-C", path, "diff", "--cached",
		"--no-ext-diff", "--quiet", "--exit-code")
}

type UpstreamStatus struct {
	Ahead  uint
	Behind uint
}

func (status UpstreamStatus) OnlyBehind() bool {
	if status.Ahead == 0 && status.Behind > 0 {
		return true
	}

	return false
}

// UpStatus returns the upstream status of the local repository.
func UpStatus(ctx context.Context, path string) (UpstreamStatus, error) {
	var status UpstreamStatus

	output, err := Out(ctx, "-C", path, "rev-list",
		"--left-right", "@{upstream}...HEAD")

	if err != nil {
		return status, err
	}

	var seeking bool

	for n := 0; n < len(output); n++ {
		char := output[n]

		if !seeking && char != '\n' {
			continue
		} else if !seeking && char == '\n' {
			seeking = true
			continue
		}

		if seeking && output[n] == '>' {
			seeking = false
			status.Ahead++
		} else if seeking && output[n] == '<' {
			seeking = false
			status.Behind++
		}
	}

	return status, nil
}
