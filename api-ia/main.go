package main

// @title           AI Ticket API
// @version         1.0
// @description     API para comunicação com modelos de IA locais através do Ollama.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Matheus
// @license.name   MIT

// @host      localhost:8080
// @BasePath  /

import (
	"log"
	"net/http"
	"os"
    _ "api-ia/docs"
    httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	srv := NewServer()

	// Carrega system prompt, se configurado.
	if path := os.Getenv("SYSTEM_PROMPT_FILE"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("não foi possível ler %s: %v", path, err)
		}

		srv.systemPrompt = string(data)
	}

	mux := http.NewServeMux()
	
	mux.Handle("GET /swagger/", httpSwagger.Handler())
	
	mux.HandleFunc("POST /api/ask", srv.handleAsk)
	mux.HandleFunc("POST /api/cancel", srv.handleCancel)

	addr := ":8080"

	log.Printf("API ouvindo em %s", addr)

	log.Fatal(http.ListenAndServe(addr, cors(mux)))
}
