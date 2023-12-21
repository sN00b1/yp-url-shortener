package handlers

import (
	"os"
)

type HandlerConfig struct {
	HandlerURL string
}

func NewHandlerConfig(urlFlag string) *HandlerConfig {
	var url string
	urlOS := os.Getenv("BASE_URL")
	if urlFlag != "" {
		url = urlFlag
	}
	if urlOS != "" {
		url = urlOS
	}
	if url == "" {
		url = "http://localhost:8080"
	}
	return &HandlerConfig{
		HandlerURL: url,
	}
}
