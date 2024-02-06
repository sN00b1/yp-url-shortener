package encoding

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (gz gzipWriter) Write(b []byte) (int, error) {
	return gz.Writer.Write(b)
}

func CompressHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(writer, request)
			return
		}

		gzwriter, err := gzip.NewWriterLevel(writer, gzip.BestSpeed)
		if err != nil {
			io.WriteString(writer, err.Error())
			return
		}
		defer gzwriter.Close()

		writer.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: writer, Writer: gzwriter}, request)
	})
}
