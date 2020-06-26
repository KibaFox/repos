package repos

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"gitlab.com/kibafox/repos/internal/errs"
)

// Parse will read the configuration file format and returns the parsed slice of
// git repositories.
func Parse(reader io.Reader, errCh chan error) ([]Repo, error) {
	repos := make([]Repo, 0, 9)

	if errCh == nil {
		return repos, errs.ErrNilChan
	}

	defer close(errCh)

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errs.ErrHomeNotFound(err)
	}

	var (
		linenum     uint
		errOccurred bool
	)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		linenum++

		str := scanner.Text()

		r, err := parseLine(home, str)
		if err != nil {
			errOccurred = true
			errCh <- fmt.Errorf("error on line %d: %w", linenum, err)

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

	if errOccurred {
		return repos, errs.ErrOccurred
	}

	return repos, nil
}

func parseLine(home, line string) (r *Repo, err error) {
	if strings.HasPrefix(line, "#") || len(line) == 0 {
		// ignore comments and empty lines
		return nil, nil
	}

	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil, nil
	} else if len(fields) != 2 {
		return nil, errs.ErrParseLine
	}

	r = &Repo{
		Path: ExpandHome(home, fields[0]),
		URL:  fields[1],
	}

	return r, nil
}
