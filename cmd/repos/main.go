package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/urfave/cli"
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
			Value: defaultConfig(),
		},
	}

	app.Commands = []cli.Command{
		ListCmd(),
		FetchCmd(),
		ImportCmd(),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func defaultConfig() (dir string) {
	if runtime.GOOS == "windows" {
		dir = filepath.Join(os.Getenv("APPDATA"), "repos")
	} else {
		dir = "~/.config/repos"
	}

	return dir
}
