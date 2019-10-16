package repos

import (
	"log"
	"os"

	"gopkg.in/src-d/go-git.v4"
)

// Sync takes a slice of git repositories and will do the equivalent of
// `git fetch` for each.  If the local repository does not exist, the
// equivalent `git clone` is performed.
func Sync(repos []Repo) error {
	for _, r := range repos {
		if _, err := os.Stat(r.Path); err == nil {
			repo, err := git.PlainOpen(r.Path)
			if err != nil {
				log.Printf("error opening repo: %s: %v", r.Path, err)
				continue
			}

			err = repo.Fetch(&git.FetchOptions{})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				log.Printf("error fetching: %s: %v", r.Path, err)
				continue
			}
		} else {
			_, err = git.PlainClone(r.Path, false, &git.CloneOptions{
				URL: r.URL,
			})
			if err != nil {
				log.Printf("error cloning: %s: %v", r.Path, err)
				continue
			}
		}
	}

	return nil
}
