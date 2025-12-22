package config

import "os"

type Config struct {
	DatabaseURL string
	RedisAddr   string
	JWTSecret   string
	GRPCPort    string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:pass@postgres:5432/ledger?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "redis:6379"),
		JWTSecret:   getEnv("JWT_SECRET", "supersecret"),
		GRPCPort:    getEnv("GRPC_PORT", ":50051"),
	}
}

func getEnv(key, fallback string) string {
	if v, exists := os.LookupEnv(key); exists {
		return v
	}
	return fallback
}
