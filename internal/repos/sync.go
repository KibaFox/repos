package repos

import (
	"context"
	"errors"
	"fmt"
	"os"

	"gopkg.in/src-d/go-git.v4"
)

var (
	ErrWorktreeNotClean = errors.New(
		"fetched, not pulled: local worktree is not clean")
	ErrUnstagedChanges = errors.New(
		"fetched, not pulled: local repo contains unstaged changes")
	ErrNonFastForwardUpdate = errors.New(
		"fetched, not pulled: changes on both remote and local")
)

// Sync takes a slice of git repositories and will do the equivalent of
// `git fetch` for each.  If the local repository does not exist, the
// equivalent `git clone` is performed.
//
// Takes in an error channel which sends errors that occur during syncing.
// The channel is closed at the end of syncing.
func Sync(ctx context.Context, repos []Repo, errs chan error) error {
	var (
		err         error
		errOccurred bool
	)

	if errs == nil {
		return errors.New("channel not initialized")
	}

	for _, r := range repos {
		if ctx.Err() != nil {
			switch ctx.Err() {
			case context.Canceled:
				return errors.New("canceled")
			case context.DeadlineExceeded:
				return errors.New("timed out")
			default:
				return fmt.Errorf("context error: %w", ctx.Err())
			}
		} else if _, err = os.Stat(r.Path); err == nil {
			err = syncExisting(ctx, r)

			if err != nil {
				errOccurred = true
				errs <- err
			}
		} else {
			err = syncNew(ctx, r)

			if err != nil {
				errOccurred = true
				errs <- err
			}
		}
	}

	close(errs)

	if errOccurred {
		return errors.New("error(s) occurred")
	}

	return nil
}

func syncNew(ctx context.Context, r Repo) (err error) {
	_, err = git.PlainCloneContext(
		ctx, r.Path, false, &git.CloneOptions{URL: r.URL})
	if err != nil {
		return fmt.Errorf("error cloning: %s: %w", r.Path, err)
	}

	return nil
}

func syncExisting(ctx context.Context, r Repo) (err error) {
	var repo *git.Repository

	repo, err = git.PlainOpen(r.Path)
	if err != nil {
		return fmt.Errorf("error opening repo: %s: %w", r.Path, err)
	}

	wrk, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %s: %w", r.Path, err)
	}

	status, err := wrk.Status()
	if err != nil {
		return fmt.Errorf("error getting worktree status: %s: %w", r.Path, err)
	}

	if !status.IsClean() {
		// only fetch, do not update
		err = repo.FetchContext(ctx, &git.FetchOptions{})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("error fetching: %s: %w", r.Path, err)
		}

		return nil
	}

	err = wrk.PullContext(ctx, &git.PullOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("error pulling: %s: %w", r.Path, convPullErr(err))
	}

	return nil
}

func convPullErr(err error) error {
	switch err {
	case git.ErrNonFastForwardUpdate:
		return ErrNonFastForwardUpdate
	case git.ErrWorktreeNotClean:
		return ErrWorktreeNotClean
	case git.ErrUnstagedChanges:
		return ErrUnstagedChanges
	default:
		return err
	}
}
