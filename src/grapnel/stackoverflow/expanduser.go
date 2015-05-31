// https://stackoverflow.com/questions/17609732/expand-tilde-to-home-directory
package stackoverflow

import (
	"os/user"
	"path/filepath"
	"strings"
)

// Returns an absolute path for 'path'; delegates to filepath.Abs for
// most of the heavy lifting. Expands '~/' in a path to the current user's
// home directory
func AbsolutePath(path string) (string, error) {
	// promote path to absolute path
	result, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	// Check in case of paths like "/something/~/something/"
	if result[:2] == "~/" {
		// attempt to get user information
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		dir := usr.HomeDir
		result = strings.Replace(result, "~/", dir, 1)
	}
	return result, nil
}
