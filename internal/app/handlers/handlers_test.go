package handlers

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sN00b1/yp-url-shortener/internal/app/config"
	"github.com/sN00b1/yp-url-shortener/internal/app/mocks"
	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		body        string
	}
	tests := []struct {
		name    string
		request string
		want    want
		body    string
		method  string
	}{
		{
			name: "post with body",
			want: want{
				contentType: "",
				statusCode:  http.StatusCreated,
				body:        "http://localhost:8080/id",
			},
			request: "/",
			method:  http.MethodPost,
			body:    "url",
		},
		{
			name: "post without body",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusInternalServerError,
				body:        "cannot generate url",
			},
			request: "/",
			method:  http.MethodPost,
			body:    "",
		},
		{
			name: "get with existing id",
			want: want{
				contentType: "text/html; charset=utf-8",
				statusCode:  http.StatusTemporaryRedirect,
				body:        "<a href=\"/url\">Temporary Redirect</a>.",
			},
			request: "/id",
			method:  http.MethodGet,
			body:    "",
		},
		{
			name: "get with null id",
			want: want{
				contentType: "",
				statusCode:  http.StatusMethodNotAllowed,
				body:        "",
			},
			request: "/",
			method:  http.MethodGet,
			body:    "",
		},
		{
			name: "get with missing id",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				body:        "cant find url by hash",
			},
			request: "/missing",
			method:  http.MethodGet,
			body:    "",
		},
		{
			name: "not supported method",
			want: want{
				contentType: "",
				statusCode:  http.StatusMethodNotAllowed,
				body:        "",
			},
			request: "/asd",
			method:  http.MethodDelete,
			body:    "",
		},
	}

	mockStorage := new(mocks.MockStorage)
	mockStorage.On("Get", "id").Return("url", nil)
	mockStorage.On("Get", "").Return("", nil)
	mockStorage.On("Get", "missing").Return("", nil)
	mockStorage.On("Get", "error").Return("", errors.New("error text"))

	mockGenerator := new(mocks.MockGenerator)
	mockGenerator.On("MakeHash", "url").Return("id", nil)
	mockGenerator.On("MakeHash", "").Return("", errors.New("err"))
	mockGenerator.On("MakeHash", "error_on_shortening").Return("", errors.New("err"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.On("Save", tt.body, "id").Return(nil)

			request := httptest.NewRequest(tt.method, tt.request, strings.NewReader(tt.body))
			writer := httptest.NewRecorder()
			cfg := config.NewConfig("", "")
			handler := NewHandler(mockStorage, mockGenerator, cfg)
			r := NewRouter(handler)
			r.ServeHTTP(writer, request)
			result := writer.Result()

			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			resBody, _ := io.ReadAll(result.Body)
			assert.Equal(t, tt.want.body, string(bytes.TrimSpace(resBody)))
		})
	}
}
