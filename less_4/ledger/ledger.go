package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type Transaction struct {
	Category string
	Amount   float64
}

type Budget struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period,omitempty"`
}

type Ledger struct {
	transactions []Transaction
	budgets      map[string]Budget
}


func NewLedger() *Ledger {
	return &Ledger{
		transactions: []Transaction{},
		budgets:      make(map[string]Budget),
	}
}

func (l *Ledger) SetBudget(b Budget) error {
	if err := b.Validate(); err != nil {
		return fmt.Errorf("budget validation failed: %w", err)
	}
	l.budgets[b.Category] = b
	return nil
}

func (l *Ledger) AddTransaction(tx Transaction) error {
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("transaction validation failed: %w", err)
	}

	b, ok := l.budgets[tx.Category]
	if !ok {
		l.transactions = append(l.transactions, tx)
		return nil
	}

	var total float64
	for _, t := range l.transactions {
		if t.Category == tx.Category {
			total += t.Amount
		}
	}

	if total+tx.Amount > b.Limit {
		return errors.New("budget exceeded")
	}

	l.transactions = append(l.transactions, tx)
	return nil
}

func (l *Ledger) LoadBudgets(r io.Reader) error {
	var budgets []Budget
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&budgets); err != nil {
		return fmt.Errorf("не удалось распарсить JSON: %w", err)
	}

	for _, b := range budgets {
		if err := l.SetBudget(b); err != nil {
			return fmt.Errorf("ошибка при установке бюджета %s: %w", b.Category, err)
		}
	}
	return nil
}

func (l *Ledger) PrintTransactions() {
	fmt.Println("=== Транзакции ===")
	for _, t := range l.transactions {
		fmt.Printf("Категория: %-10s | Сумма: %.2f\n", t.Category, t.Amount)
	}
}