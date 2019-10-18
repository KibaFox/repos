package repos

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4"
)

// FromPath will search a path for git repositories.  It builds a slice of repos
// from the paths and using the first URL for the `origin` remote configured.
func FromPath(path string) (repos []Repo, err error) {
	path = ExpandHome(path)
	repos = make([]Repo, 0, 9)

	err = filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				err = fmt.Errorf("error visiting path: %s: %w", path, err)
				log.Println(err)

				return err
			}

			if info.IsDir() && info.Name() == ".git" {
				p := filepath.Dir(path)

				repo, err := git.PlainOpen(p)
				if err != nil {
					err = fmt.Errorf("error opening repo at: %s: %w", path, err)
					log.Println(err)
					return nil
				}

				cfg, err := repo.Config()
				if err != nil {
					err = fmt.Errorf(
						"error opening git config at: %s: %w", path, err)
					log.Println(err)
					return nil
				}

				origin, ok := cfg.Remotes["origin"]
				if !ok {
					err = fmt.Errorf(
						"repo %s does not have remote: origin", path)
					log.Println(err)
					return nil
				}

				if len(origin.URLs) < 1 {
					err = fmt.Errorf(
						"repo %s has no URL for remote: origin", path)
					log.Println(err)
					return nil
				}

				r := &Repo{
					Path: p,
					URL:  origin.URLs[0],
				}

				repos = append(repos, *r)

				return filepath.SkipDir
			}

			return nil
		})
	if err != nil {
		return repos, fmt.Errorf("error walking path %s: %w", path, err)
	}

	return repos, nil
}

// WriteRepos writes the given repos in a format compatible with the parser.
func WriteRepos(repos []Repo, writer io.Writer) (err error) {
	for _, repo := range repos {
		str := fmt.Sprintf("%s %s\n", ContractHome(repo.Path), repo.URL)

		_, err = writer.Write([]byte(str))
		if err != nil {
			return fmt.Errorf("error writing repo: %s: %w", repo.Path, err)
		}
	}

	return nil
}
