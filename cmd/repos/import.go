package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/KibaFox/repos/internal/repos"
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
			var (
				file *os.File
				err  error
			)
			file, err = os.OpenFile(
				repos.ExpandHome(ImportOut), os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf(
					"import: failed to open file to write: %w", err)
			}

			defer file.Close()

			output = file
		}

		paths := args
		for i, path := range paths {
			r, err := repos.FromPath(path)
			if err != nil {
				return fmt.Errorf("import: %w", err)
			}

			_, err = output.Write(
				[]byte(fmt.Sprintf(
					"# Imported Repositories from: %s\n\n", path)))
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
	},
}
