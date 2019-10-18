package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/urfave/cli"
	"gitlab.com/KibaFox/repos/internal/repos"
)

func main() {
	app := cli.NewApp()
	app.Name = "repos"
	app.Usage = "manage multiple git repositories"

	app.Commands = []cli.Command{
		{
			Name:    "sync",
			Aliases: []string{},
			Usage:   "clone or fetch git repos",
			Action: func(c *cli.Context) error {
				return sync(c.Args().First())
			},
		},
		{
			Name:    "import",
			Aliases: []string{},
			Usage:   "walk paths with existing repos to create a configuration",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "output, o"},
			},
			Action: func(c *cli.Context) error {
				return imp(c.String("output"), c.Args()...)
			},
		},
	}

	err := app.Run(os.Args)
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

func imp(out string, paths ...string) (err error) {
	var output io.Writer

	if out == "" {
		output = os.Stdout
	} else {
		var file *os.File
		file, err = os.OpenFile(
			repos.ExpandHome(out), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("import: failed to open file to write: %w", err)
		}

		defer file.Close()

		output = file
	}

	for i, path := range paths {
		r, err := repos.FromPath(path)
		if err != nil {
			return fmt.Errorf("import: %w", err)
		}

		_, err = output.Write(
			[]byte(fmt.Sprintf("# Imported Repositories from: %s\n\n", path)))
		if err != nil {
			return fmt.Errorf("import: failed to write comment: %w", err)
		}

		err = repos.WriteRepos(r, output)
		if err != nil {
			return fmt.Errorf("import: %w", err)
		}

		if i < len(paths)-1 {
			_, err = output.Write([]byte{'\n'})
			if err != nil {
				return fmt.Errorf(
					"import: failed to write last newline: %w", err)
			}
		}
	}

	return nil
}
