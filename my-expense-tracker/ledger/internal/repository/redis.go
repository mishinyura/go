package repository

import (
	"context"
	"encoding/json"
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
		return nil, err
	}
	var data map[string]float64
	json.Unmarshal([]byte(val), &data)
	return data, nil
}

func (r *RedisRepo) SetReport(ctx context.Context, userID int64, data map[string]float64) {
	bytes, _ := json.Marshal(data)
	r.client.Set(ctx, fmt.Sprintf("report:%d", userID), bytes, 30*time.Second)
}

func (r *RedisRepo) InvalidateReport(ctx context.Context, userID int64) {
	r.client.Del(ctx, fmt.Sprintf("report:%d", userID))
}

func (r *RedisRepo) GetBudgets(ctx context.Context, userID int64) ([]*domain.Budget, error) {
	val, err := r.client.Get(ctx, fmt.Sprintf("budgets:%d", userID)).Result()
	if err != nil {
		return nil, err
	}
	var budgets []*domain.Budget
	json.Unmarshal([]byte(val), &budgets)
	return budgets, nil
}

func (r *RedisRepo) SetBudgets(ctx context.Context, userID int64, list []*domain.Budget) {
	bytes, _ := json.Marshal(list)
	r.client.Set(ctx, fmt.Sprintf("budgets:%d", userID), bytes, 30*time.Second)
}

func (r *RedisRepo) InvalidateBudgets(ctx context.Context, userID int64) {
	r.client.Del(ctx, fmt.Sprintf("budgets:%d", userID))
}
