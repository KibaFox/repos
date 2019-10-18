package repos

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
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

	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil, nil
	} else if len(fields) != 2 {
		return nil, errors.New("repos file needs to formatted: PATH REMOTE")
	}

	r = &Repo{
		Path: ExpandHome(fields[0]),
		URL:  fields[1],
	}

	return r, nil
}
