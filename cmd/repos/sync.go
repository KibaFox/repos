package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/KibaFox/repos/internal/repos"
)

var SyncFile string // nolint: gochecknoglobals

func init() { // nolint: gochecknoinits
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringVarP(&SyncFile, "file", "f", "",
		"configuration file path (default: stdin)")
}

var syncCmd = &cobra.Command{ // nolint: gochecknoglobals
	Use:   "sync",
	Short: "sync repos from a configuration",
	Long: strings.TrimSpace(`
sync will clone, pull, or fetch changes for all the git repositories listed in
the given configuration.

'git clone' is performed when the local repository does not exist or is empty.

'git pull' is performed when the local repository exists and there will be no
conflicts to update the local working diretory to the latest changes.

'git fetch' is performed when the local repository exists and there are
are potential conflicts to updating the local working directory state.

By default, the configuration is read from standard input (stdin).  You can read
from a file with the -f/--file flag.
`),
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var input io.Reader

		if SyncFile == "" {
			input = os.Stdin
		} else {
			f, err := os.Open(SyncFile)

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
	},
}
