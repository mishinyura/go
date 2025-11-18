package service

import (
	"context"
	"errors"
	"time"

	"github.com/you/monorepo/ledger/internal/domain"
)

// Ошибки уровня сервиса
var ErrBudgetExceeded = errors.New("budget exceeded")

// Интерфейс приложения, которым пользуется Gateway
type Service interface {
	CreateOrUpdateBudget(ctx context.Context, b domain.Budget) error
	ListBudgets(ctx context.Context) ([]domain.Budget, error)

	CreateTransaction(ctx context.Context, tx domain.Transaction) (int, error) // возвращает id
	ListTransactions(ctx context.Context) ([]domain.Transaction, error)

	// Дополнительно: отчет
	GetReportSummary(ctx context.Context, from, to time.Time) ([]domain.CategorySummary, error)
}

// Реализация сервиса
type service struct {
	budRepo domain.BudgetRepository
	expRepo domain.ExpenseRepository

	// для отчета можно инжектировать кэш-клиент; упростим и будем вызывать domain-функции/внешние
	reporter Reporter
}

// Reporter - интерфейс опционального слоя отчета (может использовать кэш)
type Reporter interface {
	GetSummary(ctx context.Context, from, to time.Time) ([]domain.CategorySummary, error)
}

func NewService(b domain.BudgetRepository, e domain.ExpenseRepository, r Reporter) Service {
	return &service{
		budRepo:  b,
		expRepo:  e,
		reporter: r,
	}
}

func (s *service) CreateOrUpdateBudget(ctx context.Context, b domain.Budget) error {
	if err := b.Validate(); err != nil {
		return err
	}
	return s.budRepo.Upsert(b)
}

func (s *service) ListBudgets(ctx context.Context) ([]domain.Budget, error) {
	return s.budRepo.List()
}

func (s *service) CreateTransaction(ctx context.Context, tx domain.Transaction) (int, error) {
	if err := tx.Validate(); err != nil {
		return 0, err
	}

	// Проверка бюджета через репозитории
	b, found, err := s.budRepo.GetByCategory(tx.Category)
	if err != nil {
		return 0, err
	}

	if found {
		spent, err := s.expRepo.SumByCategory(tx.Category)
		if err != nil {
			return 0, err
		}
		if spent+tx.Amount > b.Limit {
			return 0, ErrBudgetExceeded
		}
	}

	// добавляем транзакцию
	id, err := s.expRepo.Add(tx)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *service) ListTransactions(ctx context.Context) ([]domain.Transaction, error) {
	return s.expRepo.List()
}

func (s *service) GetReportSummary(ctx context.Context, from, to time.Time) ([]domain.CategorySummary, error) {
	if s.reporter != nil {
		return s.reporter.GetSummary(ctx, from, to)
	}
	return nil, nil
}
