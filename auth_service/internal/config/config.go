package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   Server
	Database Database
	Redis    Redis
	JWT      JWT
}

type Server struct {
	Port     string
	GRPCPort string
}

type Database struct {
	URL string
}

type Redis struct {
	URL string
}

type JWT struct {
	Secret           string
	AccessTTL        time.Duration
	RefreshTTL       time.Duration
	AccessTTLMinutes int
	RefreshTTLDays   int
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	accessMin := getEnvInt("JWT_ACCESS_TTL_MINUTES", 1440) // 24h для удобства тестов
	refreshDays := getEnvInt("JWT_REFRESH_TTL_DAYS", 7)

	cfg := &Config{
		Server: Server{
			Port:     getEnv("SERVER_PORT", "8081"),
			GRPCPort: getEnv("GRPC_PORT", "50053"),
		},
		Database: Database{
			URL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/auth?sslmode=disable"),
		},
		Redis: Redis{
			URL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
		},
		JWT: JWT{
			Secret:           getEnv("JWT_SECRET", "change-me-in-production"),
			AccessTTLMinutes: accessMin,
			RefreshTTLDays:   refreshDays,
			AccessTTL:        time.Duration(accessMin) * time.Minute,
			RefreshTTL:       time.Duration(refreshDays) * 24 * time.Hour,
		},
	}

	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET must be set")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
