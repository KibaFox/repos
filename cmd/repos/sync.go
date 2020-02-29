package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
	"gitlab.com/KibaFox/repos/internal/repos"
)

func SyncCmd() cli.Command {
	return cli.Command{
		Name:    "sync",
		Aliases: []string{},
		Usage:   "sync repos from a config",
		Action: func(c *cli.Context) error {
			return Sync(c.GlobalString("config"), c.Args().First())
		},
	}
}

func Sync(cfg, name string) error {
	path := filepath.Join(repos.ExpandHome(cfg), name+".repo")

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("sync: failed to open repos file: %w", err)
	}
	defer file.Close()

	errs := make(chan error, 1)

	go func() {
		for err := range errs {
			log.Println(fmt.Errorf("parse: %w", err))
		}
	}()

	r, err := repos.Parse(file, errs)
	if err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	errs = make(chan error, 1)

	go func() {
		for err := range errs {
			log.Println(fmt.Errorf("sync: %w", err))
		}
	}()

	err = repos.Sync(context.TODO(), r, errs)
	if err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}
