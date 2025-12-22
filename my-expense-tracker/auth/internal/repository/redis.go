package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) *RedisRepo {
	return &RedisRepo{client: client}
}

func (r *RedisRepo) AddToBlacklist(ctx context.Context, token string, expiration time.Duration) error {
	return r.client.Set(ctx, "blacklist:"+token, "revoked", expiration).Err()
}

func (r *RedisRepo) IsBlacklisted(ctx context.Context, token string) bool {
	exists, _ := r.client.Exists(ctx, "blacklist:"+token).Result()
	return exists > 0
}
