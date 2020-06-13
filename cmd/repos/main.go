package main

import (
	"log"
	"os"
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

   By default, configuration is read from Stdin.  You can specify a config file
   to be used instead with the '--file' or '-f' flag.

   Configurations can either be hand crafted or imported with the "import"
   subcommand.
	`)

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Usage: "configuration file path",
		},
	}

	app.Commands = []cli.Command{
		ListCmd(),
		SyncCmd(),
		ImportCmd(),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
