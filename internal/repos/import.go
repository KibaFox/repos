package repos

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/kibafox/repos/internal/errs"
	"gitlab.com/kibafox/repos/internal/git"
)

// FromPath will search a path for git repositories.  It builds a slice of repos
// from the paths and using the first URL for the `origin` remote configured.
func FromPath(
	ctx context.Context,
	path string,
	errCh chan error,
) ([]Repo, error) {
	if errCh == nil {
		return nil, errs.ErrNilChan
	}

	defer close(errCh)

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errs.ErrHomeNotFound(err)
	}

	exPath := ExpandHome(home, path)
	repos := make([]Repo, 0, 9)

	var errOccurred bool

	if err := filepath.Walk(exPath,
		func(p string, info os.FileInfo, e error) error {
			if ctx.Err() != nil {
				switch ctx.Err() {
				case context.Canceled:
					return errs.ErrContextCanceled
				case context.DeadlineExceeded:
					return errs.ErrContextTimeout
				default:
					return fmt.Errorf("context error: %w", ctx.Err())
				}
			}

			if e != nil {
				errOccurred = true
				errCh <- fmt.Errorf("error visiting path: %s: %w", p, e)
				return filepath.SkipDir
			}

			if info.IsDir() && info.Name() == ".git" {
				r := filepath.Dir(p)

				// Ignore errors here.  An empty string means we could not find
				// the remote origin.
				remote, _ := git.Origin(ctx, r)

				repos = append(repos, Repo{Path: r, URL: remote})

				return filepath.SkipDir
			}

			return nil
		}); err != nil {
		return repos, fmt.Errorf("error walking path %s: %w", path, err)
	}

	if errOccurred {
		return repos, errs.ErrOccurred
	}

	return repos, nil
}

// WriteRepos writes the given repos in a format compatible with the parser.
func WriteRepos(repos []Repo, writer io.Writer) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return errs.ErrHomeNotFound(err)
	}

	var max int
	for _, repo := range repos {
		if len(repo.Path) > max {
			max = len(repo.Path)
		}
	}

	for _, repo := range repos {
		pad := max - len(repo.Path) + 1

		var str string
		if repo.URL == "" {
			str = fmt.Sprintf(
				"# Could not find remote origin for local repository: %s\n",
				ContractHome(home, repo.Path))
		} else {
			str = fmt.Sprintf("%s%s%s\n",
				ContractHome(home, repo.Path),
				strings.Repeat(" ", pad),
				repo.URL)
		}

		_, err := writer.Write([]byte(str))
		if err != nil {
			return fmt.Errorf("error writing repo: %s: %w", repo.Path, err)
		}
	}

	return nil
}
