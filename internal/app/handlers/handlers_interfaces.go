package handlers

import "github.com/sN00b1/yp-url-shortener/internal/app/storage"

type Repository interface {
	Save(url, hash string) error
	Get(hash string) (string, error)
	Ping() error
	SaveBatchURLs(toSave []storage.ShortenURL) error
}

type Generator interface {
	MakeHash(s string) (string, error)
}
