package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/urfave/cli"
	"gitlab.com/KibaFox/repos/internal/repos"
)

func SyncCmd() cli.Command {
	return cli.Command{
		Name:    "sync",
		Aliases: []string{},
		Usage:   "sync repos from a configuration",
		Action: func(c *cli.Context) error {
			return Sync(c.GlobalString("file"))
		},
	}
}

func Sync(file string) error {
	var input io.Reader

	if file == "" {
		input = os.Stdin
	} else {
		f, err := os.Open(file)

		if err != nil {
			return fmt.Errorf("sync: failed to open repos file: %w", err)
		}
		defer f.Close()

		input = f
	}

	errs := make(chan error, 1)

	go func() {
		for err := range errs {
			log.Println(fmt.Errorf("parse: %w", err))
		}
	}()

	r, err := repos.Parse(input, errs)
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
