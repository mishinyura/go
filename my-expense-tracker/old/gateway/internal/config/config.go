package config

import (
	"os"
	"time"
)

type HTTPConfig struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type GRPCConfig struct {
	Address string
}

type Config struct {
	HTTP HTTPConfig
	GRPC GRPCConfig
}

func Load() Config {
	return Config{
		HTTP: HTTPConfig{
			Address:         getEnv("HTTP_ADDRESS", ":8081"),
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		},
		GRPC: GRPCConfig{
			Address: getEnv("GRPC_ADDRESS", "127.0.0.1:9090"),
		},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
