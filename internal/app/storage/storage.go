package storage

import (
	"errors"
	"log"
	"sync"

	"github.com/google/uuid"
)

type RamFileStorage struct {
	ramStorage  map[string]string
	fileStorage *FileStorage
	mutex       sync.RWMutex
	cfg         StorageConfig
}

func NewRamFileStorage(config *StorageConfig) (*RamFileStorage, error) {
	var tmp = make(map[string]string)

	fs, err := NewFileStorage(config.FilePath)
	if err != nil {
		log.Println(err.Error())
	}

	if fs.isActive {
		err = fs.ReadAllData(tmp)
		if err != nil {
			log.Println(err.Error())
		}
	}

	db, err := NewDBStorage(config.DBInfo)
	if err != nil {
		log.Println(err.Error())
	}

	if db.IsActive {
		err = db.ReadAllData(tmp)
		if err != nil {
			log.Println(err.Error())
		}
	}

	return &RamFileStorage{
		ramStorage:  tmp,
		fileStorage: fs,
		cfg:         *config,
	}, nil
}

func (storage *RamFileStorage) DeInit() {
	err := storage.fileStorage.Close()

	if err != nil {
		log.Println(err)
	}
}

func (storage *RamFileStorage) Save(url, hash string) error {
	_, ok := storage.ramStorage[hash]
	if ok {
		return errors.New("hash already used")
	}

	item := ShortenURL{
		ID:   uuid.NewString(),
		URL:  url,
		Hash: hash,
	}

	storage.mutex.RLock()
	storage.ramStorage[hash] = url
	storage.mutex.RUnlock()

	if storage.fileStorage.isActive {
		err := storage.fileStorage.SaveURL(item)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (storage *RamFileStorage) Get(hash string) (string, error) {
	storage.mutex.RLock()
	url, ok := storage.ramStorage[hash]
	storage.mutex.RUnlock()

	if !ok {
		return "", errors.New("cant find url by hash")
	}
	return url, nil
}

func (storage *RamFileStorage) Ping() error {
	return nil
}

func (storage *RamFileStorage) SaveBatchURLs(toSave []ShortenURL) error {
	for _, saveURL := range toSave {
		err := storage.Save(saveURL.URL, saveURL.Hash)
		if err != nil {
			log.Println(err.Error())
		}
	}

	return nil
}
