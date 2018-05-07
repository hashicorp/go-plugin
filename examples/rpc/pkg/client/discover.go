package client

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func Discover(name string) ([]string, error) {
	fileGlob := fmt.Sprintf("%s*", name)
	dir, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	//look for plugins in current directory and subfolders
	var files []string
	filesSub, err := filepath.Glob(filepath.Join(dir, "**/", fileGlob))
	filesCurrent, err := filepath.Glob(filepath.Join(dir, fileGlob))
	files = append(filesSub, filesCurrent...)

	//Look for files in system directories
	pathFile, err := exec.LookPath(name)
	if err != nil && len(pathFile) > 0 {
		files = append(files, pathFile)
	}

	// Test for Executable using exec.Lookpath, remove Extensions first as LookPath matches extension if included
	exeFiles := []string{}
	for _, v := range files {
		p := ""
		p, err = exec.LookPath(strings.TrimSuffix(v, filepath.Ext(v)))
		if err != nil {
			continue
		}
		exeFiles = append(exeFiles, p)

	}
	if len(files) < 1 {
		return nil, errors.New("No Plugins found")
	}

	return files, nil
}
