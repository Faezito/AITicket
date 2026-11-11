package main

import (
	"log"
	"net/http"
	"os"
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

	mux.HandleFunc("POST /api/ask", srv.handleAsk)
	mux.HandleFunc("POST /api/cancel", srv.handleCancel)

	addr := ":8080"

	log.Printf("API ouvindo em %s", addr)

	log.Fatal(http.ListenAndServe(addr, cors(mux)))
}
