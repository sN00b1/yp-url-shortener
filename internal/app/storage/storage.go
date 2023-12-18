package storage

import (
	"errors"
)

type Repository interface {
	Save(url, hash string) error
	Get(hash string) (string, error)
}

type Storage struct {
	ramStorage map[string]string
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
	storage.ramStorage[hash] = url
	return nil
}

func (storage *Storage) Get(hash string) (string, error) {
	url, ok := storage.ramStorage[hash]
	if !ok {
		return "", errors.New("cant find url by hash")
	}
	return url, nil
}
