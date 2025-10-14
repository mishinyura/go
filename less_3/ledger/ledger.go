package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// ---------- Структуры ----------

// Транзакция
type Transaction struct {
	Category string
	Amount   float64
}

// Бюджет
type Budget struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period,omitempty"`
}

// Ledger — хранилище транзакций и бюджетов
type Ledger struct {
	transactions []Transaction
	budgets      map[string]Budget
}

// ---------- Методы ----------

// Создание нового Ledger
func NewLedger() *Ledger {
	return &Ledger{
		transactions: []Transaction{},
		budgets:      make(map[string]Budget),
	}
}

// Установка или обновление бюджета
func (l *Ledger) SetBudget(b Budget) {
	l.budgets[b.Category] = b
}

// Добавление транзакции с проверкой бюджета
func (l *Ledger) AddTransaction(tx Transaction) error {
	// Проверяем, есть ли бюджет для категории
	b, ok := l.budgets[tx.Category]
	if !ok {
		// если нет бюджета — просто добавляем
		l.transactions = append(l.transactions, tx)
		return nil
	}

	// Считаем сумму по категории
	var total float64
	for _, t := range l.transactions {
		if t.Category == tx.Category {
			total += t.Amount
		}
	}

	// Проверяем лимит
	if total+tx.Amount > b.Limit {
		return errors.New("budget exceeded")
	}

	// Добавляем транзакцию
	l.transactions = append(l.transactions, tx)
	return nil
}

// Загрузка бюджетов из JSON
func (l *Ledger) LoadBudgets(r io.Reader) error {
	var budgets []Budget
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&budgets); err != nil {
		return fmt.Errorf("не удалось распарсить JSON: %w", err)
	}

	for _, b := range budgets {
		l.SetBudget(b)
	}
	return nil
}

// Печать всех транзакций
func (l *Ledger) PrintTransactions() {
	fmt.Println("=== Транзакции ===")
	for _, t := range l.transactions {
		fmt.Printf("Категория: %-10s | Сумма: %.2f\n", t.Category, t.Amount)
	}
}