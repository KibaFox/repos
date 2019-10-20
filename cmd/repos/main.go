package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyoh86/xdg"
	"github.com/urfave/cli"
	"gitlab.com/KibaFox/repos/internal/repos"
)

const Version = "0.0.0"

func main() {
	app := cli.NewApp()
	app.Name = "repos"
	app.Usage = "manage multiple git repositories"
	app.Version = Version
	app.Authors = []cli.Author{{Name: "Kiba Fox"}}
	app.Description = strings.TrimSpace(`

   repos uses configurations that define a list of repositories to in order to
   help you manage them.

   Configuration files can either be hand crafted or imported with the "import"
   subcommand.  Configuration file names must end in ".repo" to be used.

	`)

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "configuration directory path",
			Value: filepath.Join(xdg.ConfigHome(), "repos"),
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "list",
			Aliases: []string{},
			Usage:   "list configurations in the config directory",
			Action: func(c *cli.Context) error {
				return List(c.GlobalString("config"))
			},
		},
		{
			Name:    "fetch",
			Aliases: []string{},
			Usage:   "fetch repos from a config",
			Action: func(c *cli.Context) error {
				return Fetch(c.GlobalString("config"), c.Args().First())
			},
		},
		{
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
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func List(cfg string) error {
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
