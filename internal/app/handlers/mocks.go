package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Save(url string, id string, userID int) error {
	args := m.Called(url, id)
	return args.Error(0)
}

func (m *MockStorage) Get(id string) (storage.ShortenURL, error) {
	args := m.Called(id)
	return storage.ShortenURL{ID: uuid.NewString(), Hash: "0", URL: "http://ya.ru"}, args.Error(1)
}

func (m *MockStorage) Ping() error {
	args := m.Called()
	return args.Error(1)
}

func (m *MockStorage) SaveBatchURLs(toSace []storage.ShortenURL, userID int) error {
	args := m.Called()
	return args.Error(1)
}

func (m *MockStorage) DeInit() {
}

type MockGenerator struct {
	mock.Mock
}

func (m *MockGenerator) MakeHash(s string) (string, error) {
	args := m.Called(s)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(writer, request)
	})
}

func (m *MockStorage) ReadAllDataForUserID(ctx context.Context, userID int) ([]storage.ShortenURL, error) {
	var urls []storage.ShortenURL

	return urls, nil
}

func (m *MockStorage) DeleteByUserID(shortURLs []string, userID int) error {
	return nil
}
