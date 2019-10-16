package main

import (
	"fmt"
	"log"
	"os"

	"gitlab.com/KibaFox/repos/internal/repos"
)

func main() {
	err := sync("/home/kiba/.config/repos/personal.repo")
	if err != nil {
		log.Fatal(err)
	}
}

func sync(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("sync: failed to open repos file: %w", err)
	}
	defer file.Close()

	r, err := repos.Parse(file)
	if err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	err = repos.Sync(r)
	if err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}
