package plugin

import (
	"os"
	"path/filepath"
)

// Discover discovers plugins that are in a given directory.
//
// The directory doesn't need to be absolute. For example, "." will work fine.
//
// TODO: test
func Discover(glob, dir string) ([]string, error) {
	var err error

	// Make the directory absolute if it isn't already
	if !filepath.IsAbs(dir) {
		dir, err = filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
	}

	ls, err := filepath.Glob(filepath.Join(dir, glob))
	if err != nil {
		return nil, err
	}
	var plugins []string

	// Check for valid plugins files in glob matches
	for _, f := range ls {
		stats, err := os.Stat(f)
		if err != nil {
			return nil, err
		}
		// Skip directories
		if stats.IsDir() {
			continue
		}
		// If file is executable add to plugins
		if stats.Mode()&0111 != 0 {
			plugins = append(plugins, f)
		}

	}
	return plugins, err
}
