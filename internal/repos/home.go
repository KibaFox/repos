package repos

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func Home() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(fmt.Errorf("error getting user home directory: %w", err))
	}

	return home
}

func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		path = Home() + path[1:]
	}

	return path
}

func ContractHome(path string) string {
	h := Home()

	if strings.HasPrefix(path, h) {
		if len(path) == len(h) {
			return "~"
		}

		path = "~" + path[len(h):]
	}

	return path
}
