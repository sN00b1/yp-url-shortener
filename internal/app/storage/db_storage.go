package storage

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type DBStorage struct {
	DB *sql.DB
}

func NewDBStorage(cfg string) (*DBStorage, error) {
	objDB, err := sql.Open("postgres", cfg)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return &DBStorage{
		DB: objDB,
	}, nil
}
