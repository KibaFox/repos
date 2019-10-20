package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/urfave/cli"
	"gitlab.com/KibaFox/repos/internal/repos"
)

func ListCmd() cli.Command {
	return cli.Command{
		Name:    "list",
		Aliases: []string{},
		Usage:   "list configurations in the config directory",
		Action: func(c *cli.Context) error {
			return List(c.GlobalString("config"))
		},
	}
}

func List(cfg string) (err error) {
	path := repos.ExpandHome(cfg)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("list: could not read config dir: %w", err)
	}

	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") &&
			!file.IsDir() &&
			strings.HasSuffix(file.Name(), ".repo") {
			fmt.Println(file.Name()[0 : len(file.Name())-5])
		}
	}

	return nil
}
