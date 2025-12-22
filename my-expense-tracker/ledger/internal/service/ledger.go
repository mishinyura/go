package service

import (
	"context"
	"fmt"
	"log"

	"github.com/yuramishin/expense-tracker/ledger/internal/domain"
	"github.com/yuramishin/expense-tracker/ledger/internal/repository"
)

type LedgerService struct {
	pg    *repository.PostgresRepo
	redis *repository.RedisRepo
}

func NewLedgerService(pg *repository.PostgresRepo, redis *repository.RedisRepo) *LedgerService {
	return &LedgerService{pg: pg, redis: redis}
}

func (s *LedgerService) CreateTransaction(ctx context.Context, t *domain.Transaction) (bool, string) {
	budget, err := s.pg.GetBudget(t.UserID, t.Category)
	if err == nil && budget != nil {
		spent, _ := s.pg.GetTotalSpent(t.UserID, t.Category)
		if spent+t.Amount > budget.LimitAmount {
			msg := fmt.Sprintf("Бюджет превышен! Лимит: %.0f, Потрачено: %.0f", budget.LimitAmount, spent)
			return false, msg
		}
	}

	if err := s.pg.CreateTransaction(t); err != nil {
		return false, "DB Error"
	}

	go s.redis.InvalidateReport(context.Background(), t.UserID)

	return true, "Saved"
}

func (s *LedgerService) GetReport(ctx context.Context, userID int64) (map[string]float64, error) {
	if data, _ := s.redis.GetReport(ctx, userID); data != nil {
		log.Println("Cache HIT")
		return data, nil
	}

	log.Println("Cache MISS")
	data, err := s.pg.GetReportData(userID)
	if err != nil {
		return nil, err
	}

	go s.redis.SetReport(context.Background(), userID, data)
	return data, nil
}

func (s *LedgerService) SetBudget(ctx context.Context, userID int64, category string, limit float64) error {
	if err := s.pg.SetBudget(userID, category, limit); err != nil {
		return err
	}
	go s.redis.InvalidateBudgets(context.Background(), userID)
	return nil
}

func (s *LedgerService) GetBudgets(ctx context.Context, userID int64) ([]*domain.Budget, error) {
	if list, _ := s.redis.GetBudgets(ctx, userID); list != nil {
		log.Println("Budgets HIT")
		return list, nil
	}

	list, err := s.pg.GetBudgets(userID)
	if err != nil {
		return nil, err
	}

	go s.redis.SetBudgets(context.Background(), userID, list)
	return list, nil
}
