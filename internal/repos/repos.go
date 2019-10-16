package repos

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/src-d/go-git.v4"
)

func Sync(path string) error {
	repos, err := scanRepos(path)
	if err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	for _, r := range repos {
		if _, err := os.Stat(r.path); err == nil {
			repo, err := git.PlainOpen(r.path)
			if err != nil {
				log.Printf("sync: error opening repo: %s: %v", r.path, err)
				continue
			}

			err = repo.Fetch(&git.FetchOptions{})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				log.Printf("sync: error fetching: %s: %v", r.path, err)
				continue
			}
		} else {
			_, err = git.PlainClone(r.path, false, &git.CloneOptions{
				URL: r.url,
			})
			if err != nil {
				log.Printf("sync: error cloning: %s: %v", r.path, err)
				continue
			}
		}
	}

	return nil
}

type repo struct {
	path string
	url  string
}

func scanRepos(path string) (repos []repo, err error) {
	repos = make([]repo, 0, 9)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repos file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()

		r, err := parseLine(str)
		if err != nil {
			log.Printf("error parsing: %s: %v", str, err)
			continue
		}

		if r == nil {
			continue
		}

		repos = append(repos, *r)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("issue scanning repos file: %w", err)
	}

	return repos, nil
}

func parseLine(line string) (r *repo, err error) {
	if strings.HasPrefix(line, "#") || len(line) == 0 {
		// ignore comments and empty lines
		return nil, nil
	}

	split := strings.SplitN(line, " ", 2)
	if len(split) < 2 {
		return nil, errors.New("repos file needs to formatted: PATH REMOTE")
	}

	r = &repo{
		path: split[0],
		url:  split[1],
	}

	if strings.HasPrefix(r.path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error getting user home directory: %w", err)
		}

		r.path = home + r.path[1:]
	}

	return r, nil
}
