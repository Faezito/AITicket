package main

import (
	"api-ia/dtos"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) callOllama(ctx context.Context, req dtos.AskRequest) (string, error) {
	msgs := []dtos.OllamaMessage{}

	// System Prompt: define o comportamento da IA.
	if s.systemPrompt != "" {
		msgs = append(msgs, dtos.OllamaMessage{
			Role:    "system",
			Content: s.systemPrompt,
		})
	}

	// Contexto: fornece conhecimento adicional à IA.
	if s.context != "" {
		msgs = append(msgs, dtos.OllamaMessage{
			Role:    "system",
			Content: "Abaixo estão informações de contexto para auxiliar nas respostas:\n\n" + s.context,
		})
	}

	// Pergunta do usuário.
	msgs = append(msgs, dtos.OllamaMessage{
		Role:    "user",
		Content: req.Question,
	})

	body, err := json.Marshal(dtos.OllamaChatRequest{
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

	var out dtos.OllamaChatResponse

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
