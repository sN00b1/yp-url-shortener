package storage

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type DBStorage struct {
	DB       *sql.DB
	IsActive bool
}

func NewDBStorage(cfg string) (*DBStorage, error) {
	objDB, err := sql.Open("postgres", cfg)
	if err != nil {
		log.Println(err.Error())
		return &DBStorage{
			DB:       nil,
			IsActive: false,
		}, err
	}

	createQuery := `
		CREATE TABLE IF NOT EXIST urls (
			id SERIAL PRIMARY KEY,
			shortURL VARCHAR(255),
			originalURL VARCHAR(255),
			UNIQUE(originalURL)
		);`

	_, err = objDB.Exec(createQuery)

	if err != nil {
		log.Println(err.Error())
		return &DBStorage{
			DB:       nil,
			IsActive: false,
		}, err
	}

	return &DBStorage{
		DB:       objDB,
		IsActive: true,
	}, nil
}

func (dbStorage *DBStorage) ReadAllData(tmp map[string]string) error {
	selectAllQuery := `SELECT id, shortURL, originalURL FROM urls`

	rows, err := dbStorage.DB.Query(selectAllQuery)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var obj shortenURL
		err = rows.Scan(&obj.ID, &obj.Hash, &obj.URL)
		if err != nil {
			log.Println(err.Error())
		}
		tmp[obj.Hash] = obj.URL
	}
	return nil
}

func (dbStorage *DBStorage) SaveURL(obj shortenURL) error {
	return nil
}
