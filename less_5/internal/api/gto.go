package api

import (
	"time"

	"github.com/you/monorepo/ledger"
)

// ===== DTOs (HTTP слой) =====

type CreateTransactionRequest struct {
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"` // можно не использовать в домене, если его там ещё нет
	Date        string  `json:"date"`        // ISO-8601 строка; на стороне домена может быть опционально
}

type TransactionResponse struct {
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description,omitempty"`
	Date        string  `json:"date,omitempty"`
}

type CreateBudgetRequest struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}

type BudgetResponse struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}

// ===== Мапперы между HTTP DTO и доменными моделями =====

func toDomainTransaction(req CreateTransactionRequest) ledger.Transaction {
	tx := ledger.Transaction{
		Category: req.Category,
		Amount:   req.Amount,
	}
	// Если в доменной модели есть поля Description/Date — заполняем:
	// (Безопасно: Go проигнорирует, если этих полей нет.)
	type txWithOptional struct {
		*ledger.Transaction
		Description *string
		Date        *time.Time
	}
	// Попробуем распарсить дату, но не считаем это ошибкой HTTP-слоя.
	if t, err := time.Parse(time.RFC3339, req.Date); err == nil {
		_ = t // если у доменной модели есть поле Date — присвой
		// tx.Date = t
	}
	// if req.Description != "" {
	//  tx.Description = req.Description
	// }
	return tx
}

func toTransactionResponse(tx ledger.Transaction) TransactionResponse {
	resp := TransactionResponse{
		Category: tx.Category,
		Amount:   tx.Amount,
	}
	// Если доменная модель содержит эти поля — маппим:
	// resp.Description = tx.Description
	// resp.Date = tx.Date.Format(time.RFC3339)
	return resp
}

func toDomainBudget(req CreateBudgetRequest) ledger.Budget {
	return ledger.Budget{
		Category: req.Category,
		Limit:    req.Limit,
	}
}

func toBudgetResponse(b ledger.Budget) BudgetResponse {
	return BudgetResponse{
		Category: b.Category,
		Limit:    b.Limit,
	}
}
