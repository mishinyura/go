package cache

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init() {
	addr := getenv("REDIS_ADDR", "localhost:6379")
	pass := os.Getenv("REDIS_PASSWORD")
	dbnum, _ := strconv.Atoi(getenv("REDIS_DB", "0"))
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       dbnum,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := Client.Ping(ctx).Err(); err != nil {
		log.Fatalf("cannot connect Redis: %v", err)
	}
	log.Println("✅ connected to Redis")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
