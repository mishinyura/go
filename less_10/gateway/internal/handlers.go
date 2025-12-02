package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/you/monorepo/ledger"
)

type Handlers struct {
	Ledger *ledger.Ledger
}

func (h *Handlers) Register(mux *http.ServeMux) {
	// Prefix: /api
	mux.Handle("/api/transactions", Logger(http.HandlerFunc(h.transactions)))
	mux.Handle("/api/budgets", Logger(http.HandlerFunc(h.budgets)))
}

// /api/transactions
// POST -> создать; GET -> список
func (h *Handlers) transactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req CreateTransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		tx := toDomainTransaction(req)

		// Валидация доменной модели (из ДЗ 4)
		if err := tx.Validate(); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}

		// Добавление с проверкой бюджетов
		if err := h.Ledger.AddTransaction(tx); err != nil {
			// Предпочтительно: errors.Is(err, ledger.ErrBudgetExceeded)
			switch {
			case errors.Is(err, ledger.ErrBudgetExceeded):
				writeErr(w, http.StatusConflict, "budget exceeded")
				return
			default:
				// Фолбэк: если домен не экспортирует ErrBudgetExceeded
				if err.Error() == "budget exceeded" {
					writeErr(w, http.StatusConflict, "budget exceeded")
					return
				}
				writeErr(w, http.StatusInternalServerError, "internal error")
				return
			}
		}

		writeJSON(w, http.StatusCreated, toTransactionResponse(tx))
	case http.MethodGet:
		txs := h.Ledger.ListTransactions()
		out := make([]TransactionResponse, 0, len(txs))
		for _, t := range txs {
			out = append(out, toTransactionResponse(t))
		}
		writeJSON(w, http.StatusOK, out)
	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// /api/budgets
// POST -> создать/обновить; GET -> список
func (h *Handlers) budgets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req CreateBudgetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		b := toDomainBudget(req)

		// Валидация доменной модели (из ДЗ 4)
		if err := b.Validate(); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := h.Ledger.SetBudget(b); err != nil {
			writeErr(w, http.StatusInternalServerError, "internal error")
			return
		}
		writeJSON(w, http.StatusCreated, toBudgetResponse(b))

	case http.MethodGet:
		budgets := h.Ledger.ListBudgets()
		out := make([]BudgetResponse, 0, len(budgets))
		for _, b := range budgets {
			out = append(out, toBudgetResponse(b))
		}
		writeJSON(w, http.StatusOK, out)

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleBulkImport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var transactions []ledger.Transaction
	if err := json.NewDecoder(r.Body).Decode(&transactions); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	workers := 4
	if v := r.URL.Query().Get("workers"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			workers = n
		}
	}

	summary, err := h.ledger.BulkImport(ctx, transactions, workers)
	if err != nil {
		// Если ошибка из-за контекста — возвращаем 504 или 408
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, `{"error":"import timeout"}`, http.StatusGatewayTimeout)
			return
		}
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(summary)
}
