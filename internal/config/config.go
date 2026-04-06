package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	MaxConnLifetime time.Duration
}

type RedisConfig struct {
	URL string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", "15s"),
			WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", "15s"),
			IdleTimeout:  getDuration("SERVER_IDLE_TIMEOUT", "60s"),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/tasks"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			MaxConnLifetime: getDuration("DB_MAX_CONN_LIFETIME", "1h"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379"),
		},
	}

	if cfg.Database.URL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getDuration(key, fallback string) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		d, _ := time.ParseDuration(fallback)
		return d
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		d, _ = time.ParseDuration(fallback)
	}
	return d
}
