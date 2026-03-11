package main

import (
	"auth_service/internal/config"
	"auth_service/internal/repo/postgres"
	redisrepo "auth_service/internal/repo/redis"
	transporthttp "auth_service/internal/transport/http"
	"auth_service/internal/usecase/logic"
	postgresconn "auth_service/pkg/postgres"
	redisconn "auth_service/pkg/redis"
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := postgresconn.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		log.Error("postgres connect", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	rdb, err := redisconn.NewClient(ctx, cfg.Redis.URL)
	if err != nil {
		log.Error("redis connect", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	userRepo := postgres.NewUserRepository(pool)
	blacklistRepo := redisrepo.NewBlacklistRepository(rdb, "")
	tokenProvider := usecase.NewJWTProvider(cfg.JWT.Secret)
	authUC := usecase.NewAuthUseCase(
		userRepo,
		blacklistRepo,
		tokenProvider,
		cfg.JWT.AccessTTLMinutes,
		cfg.JWT.RefreshTTLDays,
	)
	authHandler := transporthttp.NewAuthHandler(authUC)
	e := transporthttp.NewRouter(authHandler, log)

	addr := net.JoinHostPort("", cfg.Server.Port)
	go func() {
		log.Info("server started", "port", cfg.Server.Port)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Error("server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown", "error", err)
	}
	log.Info("server stopped")
}
