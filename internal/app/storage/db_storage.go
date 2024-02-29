package storage

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
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
		CREATE TABLE IF NOT EXISTS urls (
			id VARCHAR(255) PRIMARY KEY,
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
		var obj ShortenURL
		err = rows.Scan(&obj.ID, &obj.Hash, &obj.URL)
		if err != nil {
			log.Println(err.Error())
		}
		tmp[obj.Hash] = obj.URL
	}

	err = rows.Err()
	if err != nil {
		log.Println(err.Error())
	}

	return nil
}

func (dbStorage *DBStorage) Save(url, hash string) error {
	insertSQL := `
		INSERT INTO urls (id, shortURL, originalURL)
		VALUES ($1, $2, $3)`

	_, err := dbStorage.DB.Exec(insertSQL, uuid.NewString(), hash, url)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (dbStorage *DBStorage) Get(hash string) (string, error) {
	selectSQL := `
		SELECT id, shortURL, originalURL FROM urls WHERE shortURL = $1`

	row, err := dbStorage.DB.Query(selectSQL, hash)
	if err != nil {
		return "", err
	}

	if row.Err() != nil {
		return "", row.Err()
	}

	var obj ShortenURL
	row.Next()
	err = row.Scan(&obj.ID, &obj.Hash, &obj.URL)
	if err != nil {
		return "", err
	}

	return obj.URL, nil
}

func (dbStorage *DBStorage) Ping() error {
	err := dbStorage.DB.Ping()
	return err
}

func (dbStorage *DBStorage) SaveBatchURLs(toSave []ShortenURL) error {
	tx, err := dbStorage.DB.Begin()
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	stmt, err := tx.Prepare("INSERT INTO urls(id, shortURL, originalURL) VALUES($1, $2, $3)")
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	defer stmt.Close()

	for _, saveURL := range toSave {
		_, err = stmt.Exec(
			uuid.NewString(),
			saveURL.Hash,
			saveURL.URL)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (dbStorage *DBStorage) DeInit() {
	err := dbStorage.DB.Close()

	if err != nil {
		log.Println(err.Error())
	}
}
