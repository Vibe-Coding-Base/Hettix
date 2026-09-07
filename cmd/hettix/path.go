package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// expandHome expands a leading "~/" (or "~\") in path to the current user's
// home directory. The "~user" form is not supported.
func expandHome(path string) (string, error) {
	if path == "" || path[0] != '~' {
		return path, nil
	}

	if len(path) > 1 && path[1] != '/' && path[1] != filepath.Separator {
		return "", fmt.Errorf("cannot expand user-specific home path: %q", path)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	return filepath.Join(home, path[1:]), nil
}
