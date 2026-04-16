package main

import (
	"log/slog"
	"net/http"
	"os"

	"api_gateway/internal/chat"
	"api_gateway/internal/config"
	"api_gateway/internal/middleware"
	"api_gateway/internal/proxy"
	"api_gateway/internal/user"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	e.Use(middleware.JWTValidate([]byte(cfg.JWTSecret)))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "service": "api-gateway"})
	})
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Auth service — HTTP reverse proxy
	authProxy, err := proxy.NewReverseProxy(cfg.AuthServiceURL)
	if err != nil {
		slog.Error("auth proxy", "error", err)
		os.Exit(1)
	}
	e.Any("/api/v1/auth", authProxy)
	e.Any("/api/v1/auth/*", authProxy)

	// Chat service — gRPC facade
	chatClient, err := chat.NewClient(cfg.ChatGRPCAddr)
	if err != nil {
		slog.Error("chat grpc client", "error", err)
		os.Exit(1)
	}
	defer func() { _ = chatClient.Close() }()
	chat.NewHandler(chatClient).RegisterRoutes(e)

	// User service — gRPC facade
	userClient, err := user.NewClient(cfg.UserGRPCAddr)
	if err != nil {
		slog.Error("user grpc client", "error", err)
		os.Exit(1)
	}
	defer func() { _ = userClient.Close() }()
	user.NewHandler(userClient).RegisterRoutes(e)

	// Swipe service — HTTP reverse proxy
	swipeProxy, err := proxy.NewReverseProxy(cfg.SwipeServiceURL)
	if err != nil {
		slog.Error("swipe proxy", "error", err)
		os.Exit(1)
	}
	e.Any("/v1/swipes", swipeProxy)
	e.Any("/v1/swipes/*", swipeProxy)
	e.Any("/v1/matches", swipeProxy)
	e.Any("/v1/matches/*", swipeProxy)
	e.Any("/v1/location", swipeProxy)
	e.Any("/v1/candidates", swipeProxy)

	slog.Info("gateway starting",
		"port", cfg.ServerPort,
		"auth", cfg.AuthServiceURL,
		"chat_grpc", cfg.ChatGRPCAddr,
		"user_grpc", cfg.UserGRPCAddr,
		"swipe", cfg.SwipeServiceURL,
	)
	if err := e.Start(":" + cfg.ServerPort); err != nil {
		slog.Error("gateway", "error", err)
		os.Exit(1)
	}
}
