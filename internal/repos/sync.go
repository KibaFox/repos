package repos

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/kibafox/repos/internal/errs"
	"gitlab.com/kibafox/repos/internal/git"
)

// Sync takes a slice of git repositories and will do the equivalent of
// `git fetch` for each.  If the local repository does not exist, the
// equivalent `git clone` is performed.
//
// Takes in an error channel which sends errors that occur during syncing.
// The channel is closed at the end of syncing.
func Sync(ctx context.Context, repos []Repo, errCh chan error) error {
	if errCh == nil {
		return errs.ErrNilChan
	}

	defer close(errCh)

	var errOccurred bool

	for _, r := range repos {
		if ctx.Err() != nil {
			switch ctx.Err() {
			case context.Canceled:
				return errs.ErrContextCanceled
			case context.DeadlineExceeded:
				return errs.ErrContextTimeout
			default:
				return fmt.Errorf("context error: %w", ctx.Err())
			}
		} else if _, err := os.Stat(r.Path); err == nil {
			if err := git.Pull(ctx, r.Path); err != nil {
				errOccurred = true
				errCh <- err
			}
		} else {
			if err := git.Clone(ctx, r.URL, r.Path); err != nil {
				errOccurred = true
				errCh <- err
			}
		}
	}

	if errOccurred {
		return errs.ErrOccurred
	}

	return nil
}
