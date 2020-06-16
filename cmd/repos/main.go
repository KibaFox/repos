package main

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
)

const Version = "0.0.0"

var Verbose bool // nolint: gochecknoglobals

var rootCmd = &cobra.Command{ // nolint: gochecknoglobals
	Version: Version,
	Use:     "repos",
	Short:   "manage multiple git repositories",
	Long: strings.TrimSpace(`
repos uses configurations that define a list of git repositories to in order to
help you manage them.

Configurations are given in this format:

	# Comment
	PATH URL

The PATH is the local file path of the repository.  The URL is the remote git
repository to sync from.

Configuration lines starting with '#' are ignored. Blank lines are also ignored.

Configurations can either be hand crafted or imported with the "import" command.
`),
}

func init() { // nolint: gochecknoinits
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false,
		"verbose output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
