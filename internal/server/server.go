package server

import (
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var Urls = make(map[string]string)

func ShortenerHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		getURL(writer, request)
	case http.MethodPost:
		saveURL(writer, request)
	default:
		http.Error(writer, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}
}

func saveURL(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusCreated)
	url, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	urlHash := makeHash(url)
	if urlHash == "" {
		http.Error(writer, "Cannot generate url", 500)
		return
	}
	Urls[urlHash] = string(url)
	writer.Write([]byte("http://localhost:8080/" + urlHash))
	fmt.Println("added url: ", urlHash)
}

func getURL(writer http.ResponseWriter, request *http.Request) {
	url := strings.TrimPrefix(request.URL.Path, "/")
	fmt.Println("Parsed url prefix: ", url)
	http.Redirect(writer, request, Urls[url], http.StatusTemporaryRedirect)
}

func makeHash(s []byte) string {
	h := fnv.New32a()
	_, err := h.Write(s)
	if err != nil {
		fmt.Println("makeHash error")
		return ""
	}
	return strconv.Itoa(int(h.Sum32()))
}
