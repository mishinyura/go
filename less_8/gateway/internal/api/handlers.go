package api

import (
	"net/http"

	"github.com/you/monorepo/ledger/internal/service"
)

type Handlers struct {
	svc service.Service
}

func NewHandlers(svc service.Service) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) RegisterMux() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/transactions", Logger(http.HandlerFunc(h.transactions)))
	mux.Handle("/api/budgets", Logger(http.HandlerFunc(h.budgets)))
	// другие маршруты...
	return mux
}
