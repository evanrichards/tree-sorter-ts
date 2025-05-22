package fileutil

import (
	"os"
	"path/filepath"
	"strings"
)

// HasValidExtension checks if a file has one of the valid extensions
func HasValidExtension(path string, extensions []string) bool {
	for _, ext := range extensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

// FindFiles recursively finds all files with the given extensions
func FindFiles(root string, extensions []string, recursive bool) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and node_modules
		if info.IsDir() {
			baseName := filepath.Base(path)
			if strings.HasPrefix(baseName, ".") || baseName == "node_modules" {
				return filepath.SkipDir
			}
			// Skip subdirectories if not recursive
			if !recursive && path != root {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file has valid extension
		if HasValidExtension(path, extensions) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}