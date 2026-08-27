package watcher

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func GetAllDirs(root []string) ([]string, error) {
	var dirs []string

	for _, r := range root {
		err := filepath.WalkDir(r, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("Error accessing path %s: %v\n", r, err)
				return nil
			}
			if d.IsDir() {
				dirs = append(dirs, path)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return dirs, nil
}

func IsDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}
