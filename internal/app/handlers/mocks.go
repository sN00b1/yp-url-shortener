package handlers

import (
	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Save(url string, id string) error {
	args := m.Called(url, id)
	return args.Error(0)
}

func (m *MockStorage) Get(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) Ping() error {
	args := m.Called()
	return args.Error(1)
}

func (m *MockStorage) SaveBatchURLs(toSace []storage.ShortenURL) error {
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
