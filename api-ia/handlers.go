package main

import (
	"api-ia/dtos"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// handleAsk godoc
// @Summary      Enviar pergunta para IA
// @Description  Envia uma pergunta para o modelo local através do Ollama.
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        request body dtos.AskRequest true "Dados da solicitação"
// @Success      200 {object} dtos.AskResponse "Resposta gerada com sucesso"
// @Failure      400 {object} dtos.ErrorResponse "JSON inválido ou campos obrigatórios ausentes"
// @Failure      409 {object} dtos.ErrorResponse "Request ID já está em uso"
// @Failure      499 {object} dtos.ErrorResponse "Geração cancelada"
// @Failure      502 {object} dtos.ErrorResponse "Erro na comunicação com Ollama"
// @Failure      504 {object} dtos.ErrorResponse "Tempo limite excedido"
// @Router       /api/ask [post]
func (s *Server) handleAsk(w http.ResponseWriter, r *http.Request) {
	var req dtos.AskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dtos.ErrorResponse{"JSON inválido"})
		return
	}

	if req.RequestID == "" || req.Question == "" {
		writeJSON(w, http.StatusBadRequest, dtos.ErrorResponse{
			"request_id e question são obrigatórios",
		})
		return
	}

	if req.Model == "" {
		req.Model = defaultModel
	}

	ctx, cancel := context.WithTimeout(r.Context(), askTimeout)
	defer cancel()

	if !s.registry.Add(req.RequestID, cancel) {
		writeJSON(w, http.StatusConflict, dtos.ErrorResponse{
			"request_id já em uso",
		})
		return
	}
	defer s.registry.Remove(req.RequestID)

	start := time.Now()

	answer, err := s.callOllama(ctx, req)

	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			writeJSON(w, 499, dtos.ErrorResponse{"geração cancelada"})

		case errors.Is(err, context.DeadlineExceeded):
			writeJSON(w, http.StatusGatewayTimeout, dtos.ErrorResponse{"tempo esgotado"})

		default:
			log.Printf("erro Ollama: %v", err)
			writeJSON(w, http.StatusBadGateway, dtos.ErrorResponse{err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusOK, dtos.AskResponse{
		RequestID:  req.RequestID,
		Answer:     answer,
		Model:      req.Model,
		DurationMs: time.Since(start).Milliseconds(),
	})
}

// handleCancel godoc
// @Summary      Cancelar requisição
// @Description  Cancela uma geração em andamento com base no Request ID.
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        request body dtos.CancelRequest true "Identificador da requisição"
// @Success      200 {object} map[string]string "Geração cancelada"
// @Failure      400 {object} dtos.ErrorResponse "Request ID obrigatório"
// @Failure      404 {object} dtos.ErrorResponse "Nenhuma geração em andamento com esse ID"
// @Router       /api/cancel [post]
func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	var req dtos.CancelRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RequestID == "" {
		writeJSON(w, http.StatusBadRequest, dtos.ErrorResponse{
			"request_id obrigatório",
		})
		return
	}

	if !s.registry.Cancel(req.RequestID) {
		writeJSON(w, http.StatusNotFound, dtos.ErrorResponse{
			"nenhuma geração em andamento com esse id",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "cancelado",
	})
}
