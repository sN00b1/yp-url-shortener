package storage

import (
	"errors"
	"sync"
)

type Storage struct {
	ramStorage map[string]string
	mutex      sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		ramStorage: make(map[string]string),
	}
}

func (storage *Storage) Save(url, hash string) error {
	_, ok := storage.ramStorage[hash]
	if ok {
		return errors.New("hash already used")
	}

	storage.mutex.RLock()
	storage.ramStorage[hash] = url
	storage.mutex.RUnlock()

	return nil
}

func (storage *Storage) Get(hash string) (string, error) {
	storage.mutex.RLock()
	url, ok := storage.ramStorage[hash]
	storage.mutex.RUnlock()

	if !ok {
		return "", errors.New("cant find url by hash")
	}
	return url, nil
}
