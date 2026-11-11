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
