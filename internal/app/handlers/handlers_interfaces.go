package handlers

import (
	"context"
	"net/http"

	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
)

type Repository interface {
	Save(url, hash string, userID int) error
	Get(hash string) (string, error)
	Ping() error
	SaveBatchURLs(toSave []storage.ShortenURL, userID int) error
	DeInit()
	AuthMiddleware(next http.Handler) http.Handler
	ReadAllDataForUserID(ctx context.Context, userID int) ([]storage.ShortenURL, error)
}

type Generator interface {
	MakeHash(s string) (string, error)
}
