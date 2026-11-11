
package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
	"api-ia/dtos"
)

func (s *Server) callOllama(ctx context.Context, req AskRequest) (string, error) {
    msgs := []OllamaMessage{}

    if s.systemPrompt != "" {
        msgs = append(msgs, OllamaMessage{
            Role:    "system",
            Content: s.systemPrompt,
        })
    }

    msgs = append(msgs, OllamaMessage{
        Role:    "user",
        Content: req.Question,
    })

    body, err := json.Marshal(OllamaChatRequest{
        Model:    req.Model,
        Messages: msgs,
        Stream:   false,
    })
    if err != nil {
        return "", fmt.Errorf("erro ao montar requisição: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(
        ctx,
        http.MethodPost,
        ollamaURL,
        bytes.NewReader(body),
    )
    if err != nil {
        return "", err
    }

    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := s.httpClient.Do(httpReq)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var out OllamaChatResponse

    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        return "", fmt.Errorf("resposta inválida do Ollama: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf(
            "Ollama retornou %d: %s",
            resp.StatusCode,
            out.Error,
        )
    }

    return out.Message.Content, nil
}
