package loggin

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

type (
	responseData struct {
		status int
		size   int
	}
	logginResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *logginResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *logginResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func LogginResponse(h http.Handler) http.Handler {
	logger, err := zap.NewDevelopment()
	defer logger.Sync()
	if err != nil {
		panic(err)
	}
	sugar = *logger.Sugar()
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := logginResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		uri := r.RequestURI
		method := r.Method
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
			"status", responseData.status,
			"size", responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}
