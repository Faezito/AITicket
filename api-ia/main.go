package main

// @title           AI Ticket API
// @version         1.0
// @description     API para comunicação com modelos de IA locais através do Ollama.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Matheus

// @host      localhost:8080
// @BasePath  /

import (
	_ "api-ia/docs"
	"log"
	"net/http"
	"os"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	srv := NewServer()

	// Carrega System Prompt
	systemPrompt, err := os.ReadFile("prompts/SystemPrompt.md")
	if err != nil {
		log.Fatalf("não foi possível ler SystemPrompt.md: %v", err)
	}

	srv.systemPrompt = string(systemPrompt)

	// Carrega Contexto
	context, err := os.ReadFile("context/Context.md")
	if err != nil {
		log.Fatalf("não foi possível ler Context.md: %v", err)
	}

	srv.context = string(context)

	mux := http.NewServeMux()

	mux.Handle("GET /swagger/", httpSwagger.Handler())

	mux.HandleFunc("POST /api/ask", srv.handleAsk)
	mux.HandleFunc("POST /api/cancel", srv.handleCancel)

	addr := ":8080"

	log.Printf("API ouvindo em %s", addr)

	log.Fatal(http.ListenAndServe(addr, cors(mux)))
}
