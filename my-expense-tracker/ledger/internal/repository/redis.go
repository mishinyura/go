package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yuramishin/expense-tracker/ledger/internal/domain"
)

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) *RedisRepo {
	return &RedisRepo{client: client}
}

func (r *RedisRepo) GetReport(ctx context.Context, userID int64) (map[string]float64, error) {
	val, err := r.client.Get(ctx, fmt.Sprintf("report:%d", userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var data map[string]float64
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("cache deserialization error: %w", err)
	}
	return data, nil
}

func (r *RedisRepo) SetReport(ctx context.Context, userID int64, data map[string]float64) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, fmt.Sprintf("report:%d", userID), bytes, 30*time.Second).Err()
}

func (r *RedisRepo) InvalidateReport(ctx context.Context, userID int64) error {
	return r.client.Del(ctx, fmt.Sprintf("report:%d", userID)).Err()
}

func (r *RedisRepo) GetBudgets(ctx context.Context, userID int64) ([]*domain.Budget, error) {
	val, err := r.client.Get(ctx, fmt.Sprintf("budgets:%d", userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var budgets []*domain.Budget
	if err := json.Unmarshal([]byte(val), &budgets); err != nil {
		return nil, fmt.Errorf("cache deserialization error: %w", err)
	}
	return budgets, nil
}

func (r *RedisRepo) SetBudgets(ctx context.Context, userID int64, list []*domain.Budget) error {
	bytes, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, fmt.Sprintf("budgets:%d", userID), bytes, 30*time.Second).Err()
}

func (r *RedisRepo) InvalidateBudgets(ctx context.Context, userID int64) error {
	return r.client.Del(ctx, fmt.Sprintf("budgets:%d", userID)).Err()
}
