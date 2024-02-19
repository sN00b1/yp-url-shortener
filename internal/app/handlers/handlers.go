package handlers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sN00b1/yp-url-shortener/internal/app/encoding"
	"github.com/sN00b1/yp-url-shortener/internal/app/loggin"
	"github.com/sN00b1/yp-url-shortener/internal/app/storage"
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
	CorrelationId string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type outputBatchStruct struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
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

	urlLink, err := io.ReadAll(r)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	hash, err := handler.generator.MakeHash(string(urlLink))
	if hash == "" {
		http.Error(writer, "cannot generate url", http.StatusInternalServerError)
		return
	}
	handler.storage.Save(string(urlLink), hash)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	result := fmt.Sprintf("%s/%s", handler.cfg.HandlerURL, hash)
	_, err = writer.Write([]byte(result))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func (handler *Handler) Expand(writer http.ResponseWriter, request *http.Request) {
	hash := strings.TrimPrefix(request.URL.Path, "/")
	url, err := handler.storage.Get(hash)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if url == "" {
		http.Error(writer, "cant find url by hash", http.StatusNotFound)
	}

	http.Redirect(writer, request, url, http.StatusTemporaryRedirect)
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
	if hash == "" {
		http.Error(writer, "cannot generate url", http.StatusInternalServerError)
		return
	}
	handler.storage.Save(string(input.OriginalURL), hash)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	output.Result = fmt.Sprintf("%s/%s", handler.cfg.HandlerURL, hash)

	resp, err := json.Marshal(output)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
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
		log.Println(err.Error())
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	for _, obj := range req {
		hash, err := handler.generator.MakeHash(string(obj.OriginalURL))
		if err != nil {
			log.Println(err.Error())
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		resp = append(resp, outputBatchStruct{
			CorrelationID: obj.CorrelationId,
			ShortURL:      fmt.Sprintf("%s/%s", handler.cfg.HandlerURL, hash),
		})

		toSave = append(toSave, storage.ShortenURL{
			ID:   "",
			Hash: hash,
			URL:  obj.OriginalURL,
		})
	}

	err = handler.storage.SaveBatchURLs(toSave)
	if err != nil {
		log.Println(err.Error())
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

func NewRouter(handler *Handler) chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(loggin.LogginResponse)
	router.Use(encoding.CompressHandle)
	router.Route("/", func(router chi.Router) {
		router.Get("/{id}", handler.Expand)
		router.Post("/", handler.Shorten)
		router.Post("/api/shorten", handler.ShortenFromJSON)
		router.Get("/ping", handler.Ping)
		router.Post("/api/shorten/batch", handler.PostBatchHandler)
	})
	return router
}
