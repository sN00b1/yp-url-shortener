package storage

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sN00b1/yp-url-shortener/internal/app/tools"
	"go.uber.org/zap"
)

type ShortenURL struct {
	ID     string `json:"uuid"`
	Hash   string `json:"hash"`
	URL    string `json:"url"`
	UserID int    `json:"userid"`
}

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) WriteItem(obj ShortenURL) error {
	return p.encoder.Encode(obj)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) ReadItem() (*ShortenURL, error) {
	obj := ShortenURL{}
	if err := c.decoder.Decode(&obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

type FileStorage struct {
	producer    *Producer
	consumer    *Consumer
	isActive    bool
	mutex       sync.Mutex
	usedUserIDs map[string]int
	curUserID   int
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	p, err := NewProducer(filePath)
	if err != nil {
		return &FileStorage{
			producer: nil,
			consumer: nil,
			isActive: false,
		}, err
	}

	c, err := NewConsumer(filePath)
	if err != nil {
		return &FileStorage{
			producer: nil,
			consumer: nil,
			isActive: false,
		}, err
	}

	return &FileStorage{
		producer: p,
		consumer: c,
		isActive: true,
	}, err
}

func (fileStorage *FileStorage) ReadAllData(tmp map[string]string) error {
	for {
		readItem, err := fileStorage.consumer.ReadItem()
		if err != nil {
			break
		}
		tmp[readItem.Hash] = readItem.URL
	}
	return nil
}

func (fileStorage *FileStorage) Close() error {
	err1 := fileStorage.consumer.Close()
	err2 := fileStorage.producer.Close()

	err := errors.Join(err1, err2)

	return err
}

func (fileStorage *FileStorage) SaveURL(obj ShortenURL) error {
	err := fileStorage.producer.WriteItem(obj)
	if err != nil {
		return err
	}

	return nil
}

func (fileStorage *FileStorage) IncrementID(ctx context.Context) (int, error) {
	fileStorage.curUserID++
	return fileStorage.curUserID, nil
}

func (fileStorage *FileStorage) GetLastUserID(ctx context.Context) (int, error) {
	if fileStorage != nil {
		lastUserID, err := fileStorage.IncrementID(ctx)
		if err != nil {
			log.Println("Failed to read last user id from database", zap.Error(err))
			return lastUserID, err
		}

		return lastUserID, nil
	}

	return fileStorage.curUserID, nil
}

func (fileStorage *FileStorage) SetUserIDCookie(writer http.ResponseWriter, request *http.Request, userID string) {
	claims := tools.UserClaims{
		UserID: userID,
		Claims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "myServer",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(tools.JWTSecretKey))
	if err != nil {
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	newCookie := &http.Cookie{
		Name:    tools.JWTCookieKey,
		Value:   signedToken,
		Expires: time.Now().Add(24 * time.Hour),
	}

	request.AddCookie(newCookie)

	http.SetCookie(writer, newCookie)
}

func (fileStorage *FileStorage) SaveUserID(userID string) error {
	fileStorage.mutex.Lock()

	id, err := strconv.Atoi(userID)

	if err != nil {
		return err
	}

	fileStorage.usedUserIDs[userID] = id
	fileStorage.mutex.Unlock()

	return nil
}

func (fileStorage *FileStorage) IsItCorrectUserID(userID int) bool {
	fileStorage.mutex.Lock()
	ok := fileStorage.findUserID(userID)
	fileStorage.mutex.Unlock()

	return ok
}

func (fileStorage *FileStorage) findUserID(userID int) bool {
	s := strconv.Itoa(userID)
	_, ok := fileStorage.usedUserIDs[s]
	return ok
}

func (fileStorage *FileStorage) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		cookie, err := request.Cookie(tools.JWTCookieKey)

		if err != nil && err != http.ErrNoCookie {
			log.Println("error with cookie", zap.Error(err))
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		isBatchByUserID := request.Method == http.MethodGet && request.RequestURI == "/api/user/urls"

		if err == http.ErrNoCookie {
			if isBatchByUserID {
				log.Println("No cookie and isBatchByUserID", zap.Error(err))
				http.Error(writer, "Unauthorized", http.StatusUnauthorized)
				return
			}

			lastUserID, err := fileStorage.GetLastUserID(request.Context())
			if err != nil {
				log.Println("can't get userID for cookie", zap.Error(err))
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			lastUserIDStr := strconv.Itoa(lastUserID)
			fileStorage.SetUserIDCookie(writer, request, lastUserIDStr)
			fileStorage.SaveUserID(strconv.Itoa(lastUserID))
			log.Println("Cookie is created! New user id", zap.Int("userID", lastUserID))

			next.ServeHTTP(writer, request)
		} else {
			token, userID, err := tools.GetTokenAndUserID(cookie)

			if err != nil || !token.Valid || !fileStorage.IsItCorrectUserID(userID) {
				log.Println("invalid cookie", zap.Error(err), zap.Int("userID", userID))
				http.Error(writer, "Unauthorized", http.StatusUnauthorized)
				return
			}
			log.Println("Cookie is finded", zap.Int("userID", userID))

			next.ServeHTTP(writer, request)
		}
	})
}
