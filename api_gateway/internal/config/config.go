package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	AuthGRPCAddr  string
	ChatGRPCAddr  string
	UserGRPCAddr  string
	SwipeGRPCAddr string
	JWTSecret     string
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
		ServerPort:    getEnv("GATEWAY_PORT", "8080"),
		AuthGRPCAddr:  getEnv("AUTH_GRPC_ADDR", "app:50053"),
		ChatGRPCAddr:  getEnv("CHAT_GRPC_ADDR", "chat:50051"),
		UserGRPCAddr:  getEnv("USER_GRPC_ADDR", "user:50052"),
		SwipeGRPCAddr: getEnv("SWIPE_GRPC_ADDR", "swipe:50054"),
		JWTSecret:     secret,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
