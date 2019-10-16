package repos

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"gopkg.in/src-d/go-git.v4"
)

func Sync(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open repos file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()
		if strings.HasPrefix(str, "#") || len(str) == 0 {
			// ignore comments and empty lines
			continue
		}

		split := strings.SplitN(str, " ", 2)
		if len(split) < 2 {
			return errors.New("repos file needs to formatted: PATH REMOTE")
		}

		path, remote := split[0], split[1]
		if strings.HasPrefix(path, "~/") {
			path = usrDir() + path[1:]
		}

		log.Println("Path:", path, "Remote:", remote)

		_, err = git.PlainClone(path, false, &git.CloneOptions{
			URL: remote,
		})
		if err != nil {
			return fmt.Errorf("failed to clone: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("issue scanning repos file: %w", err)
	}

	return nil
}

var home string

func usrDir() string {
	if home == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		home = usr.HomeDir
	}
	return home
}
