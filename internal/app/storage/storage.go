package storage

import (
	"errors"
	"log"
	"sync"

	"github.com/google/uuid"
)

type Storage struct {
	ramStorage map[string]string
	producer   Producer
	consumer   Consumer
	mutex      sync.RWMutex
	cfg        StorageConfig
}

func NewStorage(config *StorageConfig) (*Storage, error) {
	p, err := NewProducer(config.FilePath)
	if err != nil {
		return nil, err
	}

	c, err := NewConsumer(config.FilePath)
	if err != nil {
		return nil, err
	}

	var tmp = make(map[string]string)
	for {
		readItem, err := c.ReadItem()
		if err != nil {
			break
		}
		tmp[readItem.Hash] = readItem.URL
	}

	return &Storage{
		ramStorage: tmp,
		producer:   *p,
		consumer:   *c,
		cfg:        *config,
	}, nil
}

func (s *Storage) DeInit() {
	err1 := s.producer.Close()
	err2 := s.consumer.Close()

	var err error
	if err1 != nil && err2 != nil {
		err = errors.New(err1.Error() + " " + err2.Error())
	}
	if err1 != nil {
		err = err1
	}
	if err2 != nil {
		err = err2
	}

	if err != nil {
		log.Fatal(err)
	}
}

func (storage *Storage) Save(url, hash string) error {
	_, ok := storage.ramStorage[hash]
	if ok {
		return errors.New("hash already used")
	}

	item := shortenUrl{
		Id:   uuid.NewString(),
		URL:  url,
		Hash: hash,
	}

	if err := storage.producer.WriteItem(item); err != nil {
		log.Fatal(err)
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
