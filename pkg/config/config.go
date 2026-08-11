package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	HTTPPort    int
	GRPCPort    int
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/grpc_tasks?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "super-secret-change-in-prod"),
		HTTPPort:    getEnvInt("HTTP_PORT", 8080),
		GRPCPort:    getEnvInt("GRPC_PORT", 50051),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("invalid env var %s: %v", key, err)
		}
		return n
	}
	return fallback
}
