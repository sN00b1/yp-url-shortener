package handlers

import (
	"net/http"

	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
)

type Repository interface {
	Save(url, hash string, userID int) error
	Get(hash string) (string, error)
	Ping() error
	SaveBatchURLs(toSave []storage.ShortenURL) error
	DeInit()
	AuthMiddleware(next http.Handler) http.Handler
}

type Generator interface {
	MakeHash(s string) (string, error)
}
