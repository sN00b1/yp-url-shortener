package storage

import (
	"os"
)

type StorageConfig struct {
	FilePath string
	DBInfo   string
}

func NewStorageConfig(pathFlag, dbFlag string) *StorageConfig {
	var filePath, dbInfo string
	pathOS := os.Getenv("FILE_STORAGE_PATH")
	dbOS := os.Getenv("PSQL_SHORTENER_CONFIG")
	if pathFlag != "" {
		filePath = pathFlag
	}
	if pathOS != "" {
		filePath = pathOS
	}
	if filePath == "" {
		filePath = "/tmp/short-url-db.json"
	}

	if dbFlag != "" {
		dbInfo = dbFlag
	}
	if dbOS != "" {
		dbInfo = dbOS
	}

	return &StorageConfig{
		FilePath: filePath,
		DBInfo:   dbInfo,
	}
}
