package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort      string
	AuthServiceURL  string
	ChatGRPCAddr    string
	UserGRPCAddr    string
	SwipeServiceURL string
	JWTSecret       string // тот же секрет, что и у auth_service, для проверки токенов
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	secret := getEnv("JWT_SECRET", "")
	if secret == "" {
		secret = getEnv("GATEWAY_JWT_SECRET", "")
	}
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET or GATEWAY_JWT_SECRET must be set for gateway")
	}
	return &Config{
		ServerPort:      getEnv("GATEWAY_PORT", "8080"),
		AuthServiceURL:  getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		ChatGRPCAddr:    getEnv("CHAT_GRPC_ADDR", "chat:50051"),
		UserGRPCAddr:    getEnv("USER_GRPC_ADDR", "user:50052"),
		SwipeServiceURL: getEnv("SWIPE_SERVICE_URL", "http://swipe:8084"),
		JWTSecret:       secret,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
