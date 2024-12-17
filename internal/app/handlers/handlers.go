package handlers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sN00b1/yp-url-shortener/internal/app/encoding"
	"github.com/sN00b1/yp-url-shortener/internal/app/loggin"
	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
	"github.com/sN00b1/yp-url-shortener/internal/app/tools"
	"go.uber.org/zap"
)

type Handler struct {
	storage   Repository
	generator Generator
	mux       *chi.Mux
	cfg       HandlerConfig
}

type inputStruct struct {
	OriginalURL string `json:"url"`
}

type outputStruct struct {
	Result string `json:"result"`
}

type inputBatchStruct struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type outputBatchStruct struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type batchByUserIDResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewHandler(s Repository, g Generator, c HandlerConfig) *Handler {
	return &Handler{
		storage:   s,
		generator: g,
		mux:       chi.NewMux(),
		cfg:       c,
	}
}

func (handler *Handler) Shorten(writer http.ResponseWriter, request *http.Request) {
	r, err := decompresedReader(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	cookie, err := request.Cookie(tools.JWTCookieKey)

	if err != nil {
		loggin.Log.Debug("cannot find cookie")
		for _, cookie := range request.Cookies() {

			loggin.Log.Debug("cookie that we have", zap.String("name", cookie.Name), zap.String("Value", cookie.Value))
		}
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	token, userID, err := tools.GetTokenAndUserID(cookie)
	if err != nil || !token.Valid {
		loggin.Log.Debug("cannot find cookie", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	urlLink, err := io.ReadAll(r)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	hash, err := handler.generator.MakeHash(string(urlLink))
	if err != nil {
		http.Error(writer, "cannot generate url", http.StatusInternalServerError)
		return
	}

	headerStatus := http.StatusCreated
	err = handler.storage.Save(string(urlLink), hash, userID)
	if err != nil {
		headerStatus = http.StatusConflict
	}

	writer.WriteHeader(headerStatus)
	result := fmt.Sprintf("%s/%s", handler.cfg.HandlerURL, hash)
	_, err = writer.Write([]byte(result))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
	writer.WriteHeader(http.StatusCreated)
}

func (handler *Handler) Expand(writer http.ResponseWriter, request *http.Request) {
	hash := strings.TrimPrefix(request.URL.Path, "/")
	url, err := handler.storage.Get(hash)

	loggin.Log.Debug("Expand:", zap.String("full url", request.URL.String()))
	loggin.Log.Debug("Expand:", zap.String("find url", url))

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if url == "" {
		http.Error(writer, "cant find url by hash", http.StatusNotFound)
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.Header().Set("Location", url)
	writer.WriteHeader(http.StatusTemporaryRedirect)
}

func (handler *Handler) ShortenFromJSON(writer http.ResponseWriter, request *http.Request) {
	var input inputStruct
	var output outputStruct

	r, err := decompresedReader(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(body, &input); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := handler.generator.MakeHash(string(input.OriginalURL))
	if err != nil {
		http.Error(writer, "cannot generate url", http.StatusInternalServerError)
		return
	}

	cookie, err := request.Cookie(tools.JWTCookieKey)

	if err != nil {
		loggin.Log.Debug("cannot find cookie")
		for _, cookie := range request.Cookies() {

			loggin.Log.Debug("cookie that we have", zap.String("name", cookie.Name), zap.String("Value", cookie.Value))
		}
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	token, userID, err := tools.GetTokenAndUserID(cookie)
	if err != nil || !token.Valid {
		loggin.Log.Debug("cannot find cookie", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	headrStatus := http.StatusCreated
	err = handler.storage.Save(string(input.OriginalURL), hash, userID)
	if err != nil {
		headrStatus = http.StatusConflict
	}

	output.Result = fmt.Sprintf("%s/%s", handler.cfg.HandlerURL, hash)

	resp, err := json.Marshal(output)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(headrStatus)
	_, err = writer.Write([]byte(resp))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (handler *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	err := handler.storage.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func decompresedReader(r *http.Request) (io.Reader, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		return gzip.NewReader(r.Body)
	}
	return r.Body, nil
}

func (handler *Handler) PostBatchHandler(writer http.ResponseWriter, request *http.Request) {
	r, err := decompresedReader(request)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	var req []inputBatchStruct
	var resp []outputBatchStruct
	var toSave []storage.ShortenURL

	dec := json.NewDecoder(r)
	if err = dec.Decode(&req); err != nil {
		loggin.Log.Debug(err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	cookie, err := request.Cookie(tools.JWTCookieKey)

	if err != nil {
		loggin.Log.Debug("cannot find cookie")
		for _, cookie := range request.Cookies() {

			loggin.Log.Debug("cookie that we have", zap.String("name", cookie.Name), zap.String("Value", cookie.Value))
		}
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	token, userID, err := tools.GetTokenAndUserID(cookie)
	if err != nil || !token.Valid {
		loggin.Log.Debug("cannot find cookie", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, obj := range req {
		hash, err := handler.generator.MakeHash(string(obj.OriginalURL))
		loggin.Log.Debug("Post batch:", zap.String("hash:", hash), zap.String("OrigignalURL:", obj.OriginalURL))
		if err != nil {
			loggin.Log.Debug(err.Error())
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		resp = append(resp, outputBatchStruct{
			CorrelationID: obj.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", handler.cfg.HandlerURL, hash),
		})

		toSave = append(toSave, storage.ShortenURL{
			ID:   "",
			Hash: hash,
			URL:  obj.OriginalURL,
		})
	}

	err = handler.storage.SaveBatchURLs(toSave, userID)
	if err != nil {
		loggin.Log.Debug(err.Error())
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(writer)
	if err = enc.Encode(resp); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (handler *Handler) GetByUserIDHandler(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie(tools.JWTCookieKey)

	if err != nil {
		loggin.Log.Info("cannot find cookie", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	token, userID, err := tools.GetTokenAndUserID(cookie)
	if err != nil || !token.Valid {
		loggin.Log.Info("cannot find cookie", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	servShortURL := ""
	// так как в тестах мы не используем флаги, нужно обезопасить себя
	if handler.cfg.HandlerURL == "" {
		servShortURL = "http://localhost:8080"
	} else {
		servShortURL = handler.cfg.HandlerURL
	}

	var resp []batchByUserIDResponse
	savedURLs, err := handler.storage.ReadAllDataForUserID(request.Context(), userID)
	if err != nil {
		loggin.Log.Info("cannot read data for user", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, savedURL := range savedURLs {
		resp = append(resp, batchByUserIDResponse{
			ShortURL:    servShortURL + "/" + savedURL.Hash,
			OriginalURL: savedURL.URL,
		})
		loggin.Log.Info("Readed from batch request", zap.String("body", savedURL.URL), zap.String("result", servShortURL+"/"+savedURL.Hash), zap.Int("userID", userID))
	}

	if len(resp) == 0 {
		loggin.Log.Info("We find no urls for user", zap.Int("userID", userID))
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	loggin.Log.Info("After POST JSON request", zap.Int("count", len(resp)), zap.String("content-encoding", request.Header.Get("Content-Encoding")))

	enc := json.NewEncoder(writer)
	if err := enc.Encode(resp); err != nil {
		loggin.Log.Debug("error encoding response", zap.Error(err))
		return
	}
}

func NewRouter(handler *Handler) chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(loggin.LogginResponse)
	router.Use(encoding.CompressHandle)
	router.Use(handler.storage.AuthMiddleware)
	router.Get("/{id}", handler.Expand)
	router.Post("/", handler.Shorten)
	router.Post("/api/shorten", handler.ShortenFromJSON)
	router.Get("/ping", handler.Ping)
	router.Post("/api/shorten/batch", handler.PostBatchHandler)
	router.Get("/api/user/urls", handler.GetByUserIDHandler)
	return router
}
