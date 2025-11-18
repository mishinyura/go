package domain

import (
	"errors"
	"time"
)

// Сущности домена

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        time.Time
}

type Budget struct {
	ID       int
	Category string
	Limit    float64
}

// Валидация

func (t Transaction) Validate() error {
	if t.Category == "" {
		return errors.New("category cannot be empty")
	}
	if t.Amount == 0 {
		return errors.New("amount cannot be zero")
	}
	return nil
}

func (b Budget) Validate() error {
	if b.Category == "" {
		return errors.New("budget category cannot be empty")
	}
	if b.Limit <= 0 {
		return errors.New("budget limit must be greater than zero")
	}
	return nil
}

// Интерфейсы репозиториев - контракты для доступа к данным

// BudgetRepository - контракт для работы с бюджетами
type BudgetRepository interface {
	Upsert(b Budget) error
	GetByCategory(category string) (Budget, bool, error) // bool - found
	List() ([]Budget, error)
}

// ExpenseRepository - контракт для расходов (транзакций)
type ExpenseRepository interface {
	Add(tx Transaction) (int, error) // возвращает id
	List() ([]Transaction, error)
	SumByCategory(category string) (float64, error) // сумма по категории
}
