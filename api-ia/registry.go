package main

import (
	"context"
	"sync"
)

type Registry struct {
	mu   sync.Mutex
	jobs map[string]context.CancelFunc
}

func NewRegistry() *Registry {
	return &Registry{
		jobs: make(map[string]context.CancelFunc),
	}
}

func (r *Registry) Add(id string, cancel context.CancelFunc) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.jobs[id]; exists {
		return false
	}

	r.jobs[id] = cancel
	return true
}

func (r *Registry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.jobs, id)
}

func (r *Registry) Cancel(id string) bool {
	r.mu.Lock()
	cancel, ok := r.jobs[id]
	r.mu.Unlock()

	if ok {
		cancel()
	}

	return ok
}