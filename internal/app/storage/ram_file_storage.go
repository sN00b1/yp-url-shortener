package storage

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sN00b1/yp-url-shortener/internal/app/loggin"
	"github.com/sN00b1/yp-url-shortener/internal/app/tools"
	"go.uber.org/zap"
)

type RAMFileStorage struct {
	ramStorage     map[string]ShortenURL
	userIDStrorage map[int][]ShortenURL
	fileStorage    *FileStorage
	mutex          sync.RWMutex
	cfg            StorageConfig
	lastUserID     int
	usedUserIDs    []int
}

func NewRAMFileStorage(config *StorageConfig) (*RAMFileStorage, error) {
	var ramS = make(map[string]ShortenURL)
	var userIDS = make(map[int][]ShortenURL)

	var tmpUsers []int
	fs, err := NewFileStorage(config.FilePath)
	if err != nil {
		loggin.Log.Debug("FIle storage", zap.String("err:", err.Error()))
	}

	var lastUserID int

	if fs.isActive {
		err = fs.ReadAllData(ramS, userIDS, &tmpUsers, &lastUserID)
		if err != nil {
			loggin.Log.Debug("Read fIle storage", zap.String("err:", err.Error()))
		}
	}

	return &RAMFileStorage{
		ramStorage:     ramS,
		userIDStrorage: userIDS,
		fileStorage:    fs,
		cfg:            *config,
		lastUserID:     lastUserID,
		usedUserIDs:    tmpUsers,
	}, nil
}

func (storage *RAMFileStorage) DeInit() {
	err := storage.fileStorage.Close()

	if err != nil {
		log.Println(err)
	}
}

func (storage *RAMFileStorage) Save(url, hash string, userID int) error {
	_, ok := storage.ramStorage[hash]
	if ok {
		return errors.New("hash already used")
	}

	item := ShortenURL{
		ID:     uuid.NewString(),
		URL:    url,
		Hash:   hash,
		UserID: userID,
	}

	storage.mutex.RLock()
	storage.ramStorage[hash] = item
	storage.userIDStrorage[userID] = append(storage.userIDStrorage[userID], item)
	storage.mutex.RUnlock()

	if storage.fileStorage.isActive {
		err := storage.fileStorage.SaveURL(item)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (storage *RAMFileStorage) Get(hash string) (ShortenURL, error) {
	storage.mutex.RLock()
	item, ok := storage.ramStorage[hash]
	storage.mutex.RUnlock()

	if !ok {
		return ShortenURL{}, errors.New("cant find url by hash")
	}
	return item, nil
}

func (storage *RAMFileStorage) Ping() error {
	return nil
}

func (storage *RAMFileStorage) SaveBatchURLs(toSave []ShortenURL, userID int) error {
	for _, saveURL := range toSave {
		err := storage.Save(saveURL.URL, saveURL.Hash, userID)
		if err != nil {
			log.Println(err.Error())
		}
	}

	return nil
}

func (storage *RAMFileStorage) IncrementID(ctx context.Context) (int, error) {
	storage.lastUserID++
	if storage.lastUserID <= 0 {
		return 0, errors.New("userID out of range")
	}
	return storage.lastUserID, nil
}

func (storage *RAMFileStorage) GetLastUserID(ctx context.Context) (int, error) {
	if storage != nil {
		lastUserID, err := storage.IncrementID(ctx)
		if err != nil {
			log.Println(zap.Error(err))
			return lastUserID, err
		}
	}

	return storage.lastUserID, nil
}

func (storage *RAMFileStorage) SetUserIDCookie(writer http.ResponseWriter, request *http.Request, userID string) {
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

func (storage *RAMFileStorage) SaveUserID(userID int) {
	storage.mutex.Lock()
	storage.usedUserIDs = append(storage.usedUserIDs, userID)
	storage.mutex.Unlock()
}

func (storage *RAMFileStorage) IsItCorrectUserID(userID int) bool {
	storage.mutex.Lock()
	ok := storage.findUserID(userID)
	storage.mutex.Unlock()

	return ok
}

func (storage *RAMFileStorage) findUserID(userID int) bool {
	ok := false
	for _, v := range storage.usedUserIDs {
		if v == userID {
			ok = true
			break
		}
	}
	return ok
}

func (storage *RAMFileStorage) AuthMiddleware(next http.Handler) http.Handler {
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

			lastUserID, err := storage.GetLastUserID(request.Context())
			if err != nil {
				log.Println("can't get userID for cookie", zap.Error(err))
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			lastUserIDStr := strconv.Itoa(lastUserID)
			storage.SetUserIDCookie(writer, request, lastUserIDStr)
			storage.SaveUserID(lastUserID)
			log.Println("Cookie is created! New user id", zap.Int("userID", lastUserID))

			next.ServeHTTP(writer, request)
		} else {
			token, userID, err := tools.GetTokenAndUserID(cookie)

			if err != nil || !token.Valid || !storage.IsItCorrectUserID(userID) {
				log.Println("invalid cookie", zap.Error(err), zap.Int("userID", userID))
				http.Error(writer, "Unauthorized", http.StatusUnauthorized)
				return
			}
			log.Println("Cookie is finded", zap.Int("userID", userID))

			next.ServeHTTP(writer, request)
		}
	})
}
func (storage *RAMFileStorage) ReadAllDataForUserID(ctx context.Context, userID int) ([]ShortenURL, error) {
	var result []ShortenURL
	for _, item := range storage.userIDStrorage[userID] {
		result = append(result, storage.ramStorage[item.Hash])
	}
	return result, nil
}

func (storage *RAMFileStorage) DeleteByUserID(shortURLs []string, userID int) error {
	storage.mutex.Lock()
	for _, hash := range shortURLs {
		item, ok := storage.ramStorage[hash]

		if ok && item.UserID == userID {
			item.DeleteLog = true
			storage.ramStorage[hash] = item

			for i := 0; i < len(storage.userIDStrorage[userID]); i++ {
				if storage.userIDStrorage[userID][i].Hash == item.Hash {
					storage.userIDStrorage[userID][i].DeleteLog = true
					break
				}
			}
		}
	}
	storage.mutex.Unlock()

	if storage.fileStorage.isActive {
		storage.fileStorage.DeleteFile()
		for _, v := range storage.ramStorage {
			storage.fileStorage.SaveURL(v)
		}
	}

	return nil
}
