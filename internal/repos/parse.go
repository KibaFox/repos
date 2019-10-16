package repos

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// Parse will read the configuration file format and returns the parsed slice of
// git repositories.
func Parse(reader io.Reader) (repos []Repo, err error) {
	repos = make([]Repo, 0, 9)

	scanner := bufio.NewScanner(reader)
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
		return nil, fmt.Errorf("error scanning repos file: %w", err)
	}

	return repos, nil
}

func parseLine(line string) (r *Repo, err error) {
	if strings.HasPrefix(line, "#") || len(line) == 0 {
		// ignore comments and empty lines
		return nil, nil
	}

	split := strings.SplitN(line, " ", 2)
	if len(split) < 2 {
		return nil, errors.New("repos file needs to formatted: PATH REMOTE")
	}

	r = &Repo{
		Path: split[0],
		URL:  split[1],
	}

	if strings.HasPrefix(r.Path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error getting user home directory: %w", err)
		}

		r.Path = home + r.Path[1:]
	}

	return r, nil
}
