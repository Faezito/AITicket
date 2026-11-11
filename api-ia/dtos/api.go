package dtos

// Requisição para geração de resposta.
type AskRequest struct {
    RequestID string `json:"request_id"`
    Question  string `json:"question"`
    Model     string `json:"model,omitempty"`
}

// Resposta da geração.
type AskResponse struct {
    RequestID  string `json:"request_id"`
    Answer     string `json:"answer"`
    Model      string `json:"model"`
    DurationMs int64  `json:"duration_ms"`
}

// Requisição para cancelar uma geração.
type CancelRequest struct {
    RequestID string `json:"request_id"`
}

// Resposta de erro da API.
type ErrorResponse struct {
    Error string `json:"error"`
}
