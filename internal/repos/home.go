package repos

import (
	"strings"
)

// ExpandHome will replace ~/ with the home directory of the user.
func ExpandHome(home, path string) string {
	if strings.HasPrefix(path, "~/") {
		path = home + path[1:]
	}

	return path
}

// ContractHome will replace the user's home directory with a tilda(~).
func ContractHome(home, path string) string {
	if strings.HasPrefix(path, home) {
		if len(path) == len(home) {
			return "~"
		}

		path = "~" + path[len(home):]
	}

	return path
}
