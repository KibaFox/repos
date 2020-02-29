package repos

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrParseLine = errors.New("needs to formatted: PATH REMOTE")
)

// Parse will read the configuration file format and returns the parsed slice of
// git repositories.
func Parse(reader io.Reader, errs chan error) (repos []Repo, err error) {
	var (
		linenum     uint
		errOccurred bool
	)

	repos = make([]Repo, 0, 9)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		linenum++

		str := scanner.Text()

		r, err := parseLine(str)
		if err != nil {
			errOccurred = true
			errs <- fmt.Errorf("error on line %d: %w", linenum, err)

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

	close(errs)

	if errOccurred {
		return repos, errors.New("error(s) occurred")
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
		return nil, ErrParseLine
	}

	r = &Repo{
		Path: ExpandHome(fields[0]),
		URL:  fields[1],
	}

	return r, nil
}
