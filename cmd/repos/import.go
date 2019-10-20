package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli"
	"gitlab.com/KibaFox/repos/internal/repos"
)

func ImportCmd() cli.Command {
	return cli.Command{
		Name:    "import",
		Aliases: []string{},
		Usage:   "searches paths for existing repos to create a config",
		Description: strings.TrimSpace(`
   import will take each argument as a directory path and searches them for git
   repositories to produce one configuration file that can be used with the
   other subcommands.

   Repositories are detected when a path contains a directory named ".git".  The
   parent directory is then opened as a git repository to read the git config.
   The first URL for the remote named "origin" is used for the import.
			`),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "output, o",
				Usage: "destination file for the import (default: stdout)",
			},
		},
		Action: func(c *cli.Context) error {
			return Import(c.String("output"), c.Args()...)
		},
	}
}

func Import(out string, paths ...string) (err error) {
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
