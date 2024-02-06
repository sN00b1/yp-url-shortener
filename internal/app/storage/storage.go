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
	DBStore    DBStorage
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

	db, err := NewDBStorage(config.DBInfo)
	if err != nil {
		log.Println(err.Error())
	}

	return &Storage{
		ramStorage: tmp,
		producer:   *p,
		consumer:   *c,
		cfg:        *config,
		DBStore:    *db,
	}, nil
}

func (storage *Storage) DeInit() {
	err1 := storage.producer.Close()
	err2 := storage.consumer.Close()

	err := errors.Join(err1, err2)

	if err != nil {
		log.Print(err)
	}

	storage.DBStore.DB.Close()
}

func (storage *Storage) Save(url, hash string) error {
	_, ok := storage.ramStorage[hash]
	if ok {
		return errors.New("hash already used")
	}

	item := shortenURL{
		ID:   uuid.NewString(),
		URL:  url,
		Hash: hash,
	}

	if err := storage.producer.WriteItem(item); err != nil {
		log.Print(err)
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

func (storage *Storage) Ping() error {
	err := storage.DBStore.DB.Ping()
	return err
}
