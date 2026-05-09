package utils

import (
	"os"
	"path/filepath"
)

func GetStorageSize(basePath string) (int64, error) {
	var total int64
	dirs := []string{"db", "notes", "avatar", "assets"}
	for _, d := range dirs {
		dirPath := filepath.Join(basePath, d)
		size, err := dirSize(dirPath)
		if err == nil {
			total += size
		}
	}
	return total, nil
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
