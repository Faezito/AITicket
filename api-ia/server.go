package main

import (
	"net/http"
	"time"
)

const (
	ollamaURL    = "http://localhost:11434/api/chat"
	defaultModel = "qwen2.5:1.5b"
	askTimeout   = 5 * time.Minute
)

type Server struct {
	registry     *Registry
	httpClient   *http.Client
	systemPrompt string
	context      string
}

func NewServer() *Server {
	return &Server{
		registry:   NewRegistry(),
		httpClient: &http.Client{},
	}
}
