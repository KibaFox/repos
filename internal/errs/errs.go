package errs

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrContextCanceled occurs when the context has ben canceled.
	ErrContextCanceled = errors.New("canceled")

	// ErrContextTimeout occurs when the context deadline has exceded.
	ErrContextTimeout = errors.New("timed out")

	// ErrNilChan occurs when an uninitialized channel has been provided.
	ErrNilChan = errors.New("channel not initialized")

	// ErrOccurred signifies there were errors sent over the error channel.
	ErrOccurred = errors.New("error(s) occurred")

	// ErrParseLine occurs when parsing configuration that does not match the
	// correct format.
	ErrParseLine = errors.New("needs to formatted: PATH REMOTE")

	// ErrGit occurs when running git has a failure.
	ErrGit = errors.New("error running git")
)

// ErrHomeNotFound occurs when there is an error using os.UserHomeDir().
func ErrHomeNotFound(err error) error {
	return fmt.Errorf("error finding home directory: %w", err)
}

// NewErrGit creats a new Git error.
func NewErrGit(stderr fmt.Stringer, cmdArgs ...string) error {
	return fmt.Errorf("%w %s:\n%s", ErrGit,
		strings.Join(cmdArgs, " "),
		strings.TrimSuffix(stderr.String(), "\n"),
	)
}

// ErrContext occurs when there is a context error that is not a canecelation or
// a timeout.
func ErrContext(ctx context.Context) error {
	return fmt.Errorf("context error: %w", ctx.Err())
}
