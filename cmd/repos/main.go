package main

import (
	"log"

	"gitlab.com/KibaFox/repos/internal/repos"
)

func main() {
	err := repos.Sync("/home/kiba/.config/repos/personal.repo")
	if err != nil {
		log.Fatal(err)
	}
}
