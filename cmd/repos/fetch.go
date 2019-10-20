package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
	"gitlab.com/KibaFox/repos/internal/repos"
)

func FetchCmd() cli.Command {
	return cli.Command{
		Name:    "fetch",
		Aliases: []string{},
		Usage:   "fetch repos from a config",
		Action: func(c *cli.Context) error {
			return Fetch(c.GlobalString("config"), c.Args().First())
		},
	}
}

func Fetch(cfg, name string) error {
	path := filepath.Join(repos.ExpandHome(cfg), name+".repo")

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
