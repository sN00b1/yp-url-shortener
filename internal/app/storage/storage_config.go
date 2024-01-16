package storage

import (
	"os"
)

type StorageConfig struct {
	FilePath string
}

func NewStorageConfig(pathFlag string) *StorageConfig {
	var filePath string
	pathOS := os.Getenv("FILE_STORAGE_PATH")
	if pathFlag != "" {
		filePath = pathFlag
	}
	if pathOS != "" {
		filePath = pathOS
	}
	if filePath == "" {
		filePath = "/tmp/short-url-db.json"
	}
	return &StorageConfig{
		FilePath: filePath,
	}
}
