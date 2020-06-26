package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/kibafox/repos/internal/errs"
	"gitlab.com/kibafox/repos/internal/repos"
)

var ImportOut string // nolint: gochecknoglobals

func init() { // nolint: gochecknoinits
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVarP(&ImportOut, "out", "o", "",
		"destination file for the import (default: stdout)")
}

var importCmd = &cobra.Command{ // nolint: gochecknoglobals
	Use:   "import [flags] directory ...",
	Short: "searches paths for existing repos to create a config",
	Long: strings.TrimSpace(`
import searches through directories for repositiories to create a configuration
that can be used for other commands.

One or more directories can be given as arguments.  Each directory given is
traversed recursively until a git repository is found.

Repositories are detected when a path contains a directory named ".git".  The
path found will become the PATH part of the config entry.  The git repository is
inspected and the first remote named "origin" is used for the URL part of the
config entry.

By default, the configuration is written to standard output (stdout).  You can
write to a file with the -o/--out flag.
`),
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var output io.Writer

		if ImportOut == "" {
			output = os.Stdout
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return errs.ErrHomeNotFound(err)
			}

			file, err := os.OpenFile(
				repos.ExpandHome(home, ImportOut),
				os.O_CREATE|os.O_WRONLY,
				0644,
			)
			if err != nil {
				return fmt.Errorf(
					"import: failed to open file to write: %w", err)
			}

			defer file.Close()

			output = file
		}

		paths := args
		for i, path := range paths {

			errs := make(chan error, 1)

			go func() {
				for err := range errs {
					log.Println(fmt.Errorf("import: %w", err))
				}
			}()

			r, rErr := repos.FromPath(context.TODO(), path, errs)
			if rErr != nil {
				return fmt.Errorf("import: %w", rErr)
			}

			_, iErr := output.Write(
				[]byte(fmt.Sprintf(
					"# Imported Repositories from: %s\n\n", path)))
			if iErr != nil {
				return fmt.Errorf("import: failed to write comment: %w", iErr)
			}

			if wErr := repos.WriteRepos(r, output); wErr != nil {
				return fmt.Errorf("import: %w", wErr)
			}

			if i < len(paths)-1 {
				_, nlErr := output.Write([]byte{'\n'})
				if nlErr != nil {
					return fmt.Errorf(
						"import: failed to write last newline: %w", nlErr)
				}
			}
		}

		return nil
	},
}
