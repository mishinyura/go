package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	pb "github.com/yuramishin/expense-tracker/proto/pb_ledger"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(addr string) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr, DB: 0})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, err
	}
	fmt.Println("Redis Connected!")
	return &RedisClient{Client: rdb}, nil
}

func (r *RedisClient) GetReport(ctx context.Context, userID int64) (map[string]float64, error) {
	val, err := r.Client.Get(ctx, fmt.Sprintf("report:%d", userID)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var data map[string]float64
	json.Unmarshal([]byte(val), &data)
	return data, nil
}

func (r *RedisClient) SetReport(ctx context.Context, userID int64, data map[string]float64) {
	bytes, _ := json.Marshal(data)
	r.Client.Set(ctx, fmt.Sprintf("report:%d", userID), bytes, 30*time.Second)
}

func (r *RedisClient) GetBudgets(ctx context.Context, userID int64) ([]*pb.Budget, error) {
	key := fmt.Sprintf("budgets:%d", userID)
	val, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var budgets []*pb.Budget
	err = json.Unmarshal([]byte(val), &budgets)
	return budgets, err
}

func (r *RedisClient) SetBudgets(ctx context.Context, userID int64, budgets []*pb.Budget) {
	key := fmt.Sprintf("budgets:%d", userID)
	bytes, _ := json.Marshal(budgets)
	r.Client.Set(ctx, key, bytes, 30*time.Second)
}

func (r *RedisClient) InvalidateBudgets(ctx context.Context, userID int64) {
	key := fmt.Sprintf("budgets:%d", userID)
	r.Client.Del(ctx, key)
	fmt.Println("Cache INVALIDATED for budgets user", userID)
}
