package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr    string
	GRPCAddr    string
	DBDSN       string
	WorkerCount int
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:    getEnv("TODO_HTTP_ADDR", ":8080"),
		GRPCAddr:    getEnv("TODO_GRPC_ADDR", ":9090"),
		DBDSN:       getEnv("TODO_DB_DSN", "postgres://todo:todo@localhost:5432/todo?sslmode=disable"),
		WorkerCount: getEnvInt("TODO_WORKERS", 4),
	}

	if cfg.WorkerCount < 1 {
		return Config{}, fmt.Errorf("TODO_WORKERS must be >= 1")
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func getEnvInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return n
}