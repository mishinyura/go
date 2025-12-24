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
			msg := fmt.Sprintf("Budget exceeded! Limit: %.0f, Spent: %.0f", budget.LimitAmount, spent)
			return false, msg
		}
	}

	if err := s.pg.CreateTransaction(t); err != nil {
		return false, "DB Error"
	}

	go func() {
		if err := s.redis.InvalidateReport(context.Background(), t.UserID); err != nil {
			log.Printf("Redis error (InvalidateReport): %v", err)
		}
	}()

	return true, "Saved"
}

func (s *LedgerService) GetReport(ctx context.Context, userID int64) (map[string]float64, error) {
	data, err := s.redis.GetReport(ctx, userID)
	if err != nil {
		log.Printf("Redis error: %v", err)
	} else if data != nil {
		return data, nil
	}

	data, err = s.pg.GetReportData(userID)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := s.redis.SetReport(context.Background(), userID, data); err != nil {
			log.Printf("Redis error (SetReport): %v", err)
		}
	}()

	return data, nil
}

func (s *LedgerService) SetBudget(ctx context.Context, userID int64, category string, limit float64) error {
	if err := s.pg.SetBudget(userID, category, limit); err != nil {
		return err
	}

	go func() {
		if err := s.redis.InvalidateBudgets(context.Background(), userID); err != nil {
			log.Printf("Redis error (InvalidateBudgets): %v", err)
		}
	}()
	return nil
}

func (s *LedgerService) GetBudgets(ctx context.Context, userID int64) ([]*domain.Budget, error) {
	list, err := s.redis.GetBudgets(ctx, userID)
	if err != nil {
		log.Printf("Redis error: %v", err)
	} else if list != nil {
		return list, nil
	}

	list, err = s.pg.GetBudgets(userID)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := s.redis.SetBudgets(context.Background(), userID, list); err != nil {
			log.Printf("Redis error (SetBudgets): %v", err)
		}
	}()
	return list, nil
}
