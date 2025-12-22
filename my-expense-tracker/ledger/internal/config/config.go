package config

import "os"

type Config struct {
	DatabaseURL string
	RedisAddr   string
	GRPCPort    string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:pass@postgres:5432/ledger?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "redis:6379"),
		GRPCPort:    getEnv("GRPC_PORT", ":50052"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
