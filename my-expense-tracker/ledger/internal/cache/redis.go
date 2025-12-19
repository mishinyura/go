package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(addr string) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})
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
