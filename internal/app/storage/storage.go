package storage

import (
	"errors"
	"log"
	"sync"

	"github.com/google/uuid"
)

type Storage struct {
	ramStorage  map[string]string
	fileStorage *FileStorage
	mutex       sync.RWMutex
	cfg         StorageConfig
	dbStore     *DBStorage
}

func NewStorage(config *StorageConfig) (*Storage, error) {
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

	return &Storage{
		ramStorage:  tmp,
		fileStorage: fs,
		cfg:         *config,
		dbStore:     db,
	}, nil
}

func (storage *Storage) DeInit() {
	err := storage.fileStorage.Close()

	if err != nil {
		log.Println(err)
	}

	storage.dbStore.DB.Close()
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

	storage.mutex.RLock()
	storage.ramStorage[hash] = url
	storage.mutex.RUnlock()

	if storage.dbStore.IsActive {
		err := storage.dbStore.SaveURL(item)
		if err != nil {
			log.Println(err.Error())
		} else {
			return nil
		}
	}

	if storage.fileStorage.isActive {
		err := storage.fileStorage.SaveURL(item)
		if err != nil {
			log.Println(err)
		}
	}

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
	err := storage.dbStore.DB.Ping()
	return err
}
