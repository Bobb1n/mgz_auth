package main

import (
	"log/slog"
	"os"
	"net/http"
    "github.com/prometheus/client_golang/prometheus/promhttp"
	"api_gateway/internal/config"
	"api_gateway/internal/chat"
	"api_gateway/internal/middleware"
	"api_gateway/internal/proxy"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	e := echo.New()
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.CORS())
	// Проверка JWT для всех маршрутов кроме /health и /api/v1/auth; прокидывает X-User-Id, X-User-Email, X-User-Username в бэкенды
	e.Use(middleware.JWTValidate([]byte(cfg.JWTSecret)))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok", "service": "api-gateway"})
	})
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	authProxy, err := proxy.NewReverseProxy(cfg.AuthServiceURL)
	if err != nil {
		slog.Error("auth proxy", "error", err)
		os.Exit(1)
	}
	chatProxy, err := proxy.NewReverseProxy(cfg.ChatServiceURL)
	if err != nil {
		slog.Error("chat proxy", "error", err)
		os.Exit(1)
	}
	userProxy, err := proxy.NewReverseProxy(cfg.UserServiceURL)
	if err != nil {
		slog.Error("user proxy", "error", err)
		os.Exit(1)
	}

	e.Any("/api/v1/auth", authProxy)
	e.Any("/api/v1/auth/*", authProxy)
	e.Any("/v1/*", chatProxy)
	e.Any("/api/v1/users", userProxy)
	e.Any("/api/v1/users/*", userProxy)

	chatClient, err := chat.NewClient(cfg.ChatGRPCAddr)
	if err != nil {
		slog.Error("chat grpc client", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = chatClient.Close()
	}()
	chatHandler := chat.NewHandler(chatClient)
	chatHandler.RegisterRoutes(e)

	slog.Info("gateway starting",
		"port", cfg.ServerPort,
		"auth", cfg.AuthServiceURL,
		"chat_http", cfg.ChatServiceURL,
		"chat_grpc", cfg.ChatGRPCAddr,
		"user", cfg.UserServiceURL,
	)
	if err := e.Start(":" + cfg.ServerPort); err != nil {
		slog.Error("gateway", "error", err)
		os.Exit(1)
	}
}
