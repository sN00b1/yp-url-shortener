package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sN00b1/yp-url-shortener/internal/app/loggin"
	"github.com/sN00b1/yp-url-shortener/internal/app/tools"
	"go.uber.org/zap"
)

type DBStorage struct {
	DB            *sql.DB
	IsActive      bool
	lastUserID    int
	mutex         sync.RWMutex
	usedUserIDs   []int
	deleteOnce    sync.Once
	deleteJobs    chan deleteJob
	deleteWorkers int
}

type deleteJob struct {
	shortURLs []string
	userID    int
	result    chan error
}

const deleteWorkerDefault = 4
const deleteChunkSize = 100

func (dbStorage *DBStorage) initDeletePool() {
	dbStorage.deleteOnce.Do(func() {
		workerCount := deleteWorkerDefault
		stats := dbStorage.DB.Stats()
		if stats.MaxOpenConnections > 0 && stats.MaxOpenConnections < workerCount {
			workerCount = stats.MaxOpenConnections
		}
		if workerCount <= 0 {
			workerCount = 1
		}

		dbStorage.deleteWorkers = workerCount
		dbStorage.deleteJobs = make(chan deleteJob, workerCount*2)

		for i := 0; i < workerCount; i++ {
			go dbStorage.deleteWorker()
		}
	})
}

func (dbStorage *DBStorage) deleteWorker() {
	for job := range dbStorage.deleteJobs {
		err := dbStorage.processDeleteJob(job)
		job.result <- err
		close(job.result)
	}
}

func (dbStorage *DBStorage) processDeleteJob(job deleteJob) error {
	ctx := context.Background()
	query := `
		UPDATE urls
		SET deleted = TRUE
		WHERE shortURL = ANY($1)
		AND userID = $2;
	`

	chunks := tools.ChunkSlice(job.shortURLs, deleteChunkSize)
	for _, chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}

		loggin.Log.Info("Deleting URLs", zap.Strings("hashes", chunk), zap.Int("userID", job.userID))

		res, err := dbStorage.DB.ExecContext(ctx, query, pq.Array(chunk), job.userID)

		if err != nil {
			return err
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}

		loggin.Log.Info("Deleted URLs from database", zap.Int64("count", rowsAffected), zap.Int("userID", job.userID))
	}

	return nil
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
			userID INT,
			deleted BOOLEAN DEFAULT FALSE,
			UNIQUE(originalURL, userID)
		);
		CREATE TABLE IF NOT EXISTS last_user_id (
			id INT PRIMARY KEY DEFAULT 1
		);
		INSERT INTO last_user_id (id) VALUES (1) ON CONFLICT DO NOTHING;`

	_, err = objDB.Exec(createQuery)

	if err != nil {
		loggin.Log.Debug("err:", zap.String("err:", err.Error()))
		return &DBStorage{
			DB:       nil,
			IsActive: false,
		}, err
	}

	db := &DBStorage{
		DB:       objDB,
		IsActive: true,
	}
	db.ReadAllData()
	return db, nil
}

func (dbStorage *DBStorage) ReadAllData() error {
	selectAllQuery := `SELECT id, shortURL, originalURL, userID FROM urls`

	rows, err := dbStorage.DB.Query(selectAllQuery)
	if err != nil {
		return err
	}

	defer rows.Close()

	dbStorage.lastUserID = 1
	for rows.Next() {
		var obj ShortenURL
		err = rows.Scan(&obj.ID, &obj.Hash, &obj.URL, &obj.UserID)
		if err != nil {
			log.Println(err.Error())
		}
		dbStorage.lastUserID = obj.UserID
	}

	err = rows.Err()
	if err != nil {
		log.Println(err.Error())
	}

	return nil
}

func (dbStorage *DBStorage) Save(url, hash string, userID int) error {
	insertSQL := `
		INSERT INTO urls (id, shortURL, originalURL, userID)
		VALUES ($1, $2, $3, $4)`

	_, err := dbStorage.DB.Exec(insertSQL, uuid.NewString(), hash, url, userID)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (dbStorage *DBStorage) Get(hash string) (ShortenURL, error) {
	selectSQL := `
		SELECT id, shortURL, originalURL, userID FROM urls WHERE shortURL = $1`

	row, err := dbStorage.DB.Query(selectSQL, hash)
	if err != nil {
		return ShortenURL{}, err
	}
	defer row.Close()

	if row.Err() != nil {
		return ShortenURL{}, row.Err()
	}

	var obj ShortenURL
	row.Next()
	err = row.Scan(&obj.ID, &obj.Hash, &obj.URL, &obj.UserID)
	if err != nil {
		return ShortenURL{}, err
	}

	return obj, nil
}

func (dbStorage *DBStorage) Ping() error {
	if dbStorage.DB != nil {
		err := dbStorage.DB.Ping()
		return err
	}
	return errors.New("db is not init")
}

func (dbStorage *DBStorage) SaveBatchURLs(toSave []ShortenURL, userID int) error {
	tx, err := dbStorage.DB.Begin()
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO urls(id, shortURL, originalURL, userID) VALUES($1, $2, $3, $4)")
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	defer stmt.Close()

	for _, saveURL := range toSave {
		_, err = stmt.Exec(
			uuid.NewString(),
			saveURL.Hash,
			saveURL.URL,
			userID)

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

func (dbStorage *DBStorage) IncrementID(ctx context.Context) (int, error) {
	var newID int
	err := dbStorage.DB.QueryRowContext(ctx, `
		WITH updated AS (
			UPDATE last_user_id
			SET id = id + 1
			RETURNING id
		)
		SELECT id FROM updated
		UNION ALL
		SELECT id FROM last_user_id WHERE NOT EXISTS (SELECT 1 FROM updated)
	`).Scan(&newID)
	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (dbStorage *DBStorage) GetLastUserID(ctx context.Context) (int, error) {
	if dbStorage.DB != nil {
		lastUserID, err := dbStorage.IncrementID(ctx)
		if err != nil {
			log.Println("Failed to read last user id from database", zap.Error(err))
			return lastUserID, err
		}

		return lastUserID, nil
	}

	dbStorage.lastUserID = dbStorage.lastUserID + 1
	return dbStorage.lastUserID, nil
}

func (dbStorage *DBStorage) SetUserIDCookie(writer http.ResponseWriter, request *http.Request, userID string) {
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

func (dbStorage *DBStorage) SaveUserID(userID int) {
	dbStorage.mutex.Lock()
	dbStorage.usedUserIDs = append(dbStorage.usedUserIDs, userID)
	dbStorage.mutex.Unlock()
}

func (dbStorage *DBStorage) IsItCorrectUserID(userID int) bool {
	dbStorage.mutex.RLock()
	ok := dbStorage.findUserID(userID)
	dbStorage.mutex.RUnlock()

	return ok
}

func (dbStorage *DBStorage) findUserID(userID int) bool {
	ok := false
	for _, v := range dbStorage.usedUserIDs {
		if v == userID {
			ok = true
			break
		}
	}
	return ok
}

func (dbStorage *DBStorage) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		cookie, err := request.Cookie(tools.JWTCookieKey)

		if err != nil && err != http.ErrNoCookie {
			log.Println("error with cookie", zap.Error(err))
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if err == http.ErrNoCookie {
			lastUserID, err := dbStorage.GetLastUserID(request.Context())
			if err != nil {
				log.Println("can't get userID for cookie", zap.Error(err))
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			lastUserIDStr := strconv.Itoa(lastUserID)
			dbStorage.SetUserIDCookie(writer, request, lastUserIDStr)
			dbStorage.SaveUserID(lastUserID)
			log.Println("Cookie is created! New user id", zap.Int("userID", lastUserID))

			next.ServeHTTP(writer, request)
		} else {
			token, userID, err := tools.GetTokenAndUserID(cookie)

			if err != nil || !token.Valid || !dbStorage.IsItCorrectUserID(userID) {
				log.Println("invalid cookie", zap.Error(err), zap.Int("userID", userID))
				http.Error(writer, "Unauthorized", http.StatusUnauthorized)
				return
			}
			log.Println("Cookie is finded", zap.Int("userID", userID))

			next.ServeHTTP(writer, request)
		}
	})
}

func (dbStorage *DBStorage) ReadAllDataForUserID(ctx context.Context, userID int) ([]ShortenURL, error) {
	if dbStorage.DB == nil {
		return []ShortenURL{}, errors.New("data base does not connected")
	}

	urls, err := dbStorage.SelectSavedURLsForUserID(ctx, userID)
	if err != nil {
		log.Println("Failed to read from database", zap.Error(err))
		return []ShortenURL{}, err
	}

	return urls, err
}

func (dbStorage *DBStorage) SelectSavedURLsForUserID(ctx context.Context, userID int) ([]ShortenURL, error) {
	var savedURLs []ShortenURL
	var emptyURLs []ShortenURL

	sqlStatement := `SELECT id, shortURL, originalURL, userID FROM urls where userID = $1`
	rows, err := dbStorage.DB.QueryContext(ctx, sqlStatement, userID)
	if err != nil {
		log.Println("Failed to read from database", zap.Error(err))
		return emptyURLs, err
	}
	defer rows.Close()

	for rows.Next() {
		var obj ShortenURL
		err = rows.Scan(&obj.ID, &obj.Hash, &obj.URL, &obj.UserID)
		if err != nil {
			log.Println("Failed to read from database", zap.Error(err))
			return emptyURLs, err
		}
		savedURLs = append(savedURLs, obj)
	}

	err = rows.Err()
	if err != nil {
		log.Println("Failed to read from database", zap.Error(err))
		return emptyURLs, err
	}

	return savedURLs, err
}

func (dbStorage *DBStorage) DeleteByUserID(shortURLs []string, userID int) error {
	if len(shortURLs) == 0 {
		return nil
	}

	dbStorage.initDeletePool()
	resultCh := make(chan error, 1)
	job := deleteJob{
		shortURLs: append([]string(nil), shortURLs...),
		userID:    userID,
		result:    resultCh,
	}

	select {
	case dbStorage.deleteJobs <- job:
	case <-time.After(5 * time.Second):
		return errors.New("timeout enqueueing delete job")
	}

	return <-resultCh
}
