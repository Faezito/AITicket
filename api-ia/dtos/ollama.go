package dtos

// Mensagem enviada ao Ollama.
type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Requisição para o endpoint /api/chat.
type OllamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

// Resposta recebida do Ollama.
type OllamaChatResponse struct {
	Model   string        `json:"model"`
	Message OllamaMessage `json:"message"`
	Error   string        `json:"error,omitempty"`
}
