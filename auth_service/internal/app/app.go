package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"auth_service/internal/config"
	"auth_service/internal/repo/postgres"
	redisrepo "auth_service/internal/repo/redis"
	grpctransport "auth_service/internal/transport/grpc"
	httptransport "auth_service/internal/transport/http"
	"auth_service/internal/usecase"
	authv1 "auth_service/pkg/api/auth/v1"
	pgconn "auth_service/pkg/postgres"
	rdconn "auth_service/pkg/redis"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type App struct {
	cfg     *config.Config
	log     *slog.Logger
	pool    *pgxpool.Pool
	rdb     *redis.Client
	httpSrv *http.Server
	grpcSrv *grpc.Server
	grpcLis net.Listener
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*App, error) {
	pool, err := pgconn.NewPoolFromURL(ctx, cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	rdb, err := rdconn.NewFromURL(ctx, cfg.Redis.URL)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("redis: %w", err)
	}

	users := postgres.NewUserRepository(pool)
	blacklist := redisrepo.NewBlacklistRepository(rdb, "")
	tokens := usecase.NewJWTProvider(cfg.JWT.Secret)
	authUC := usecase.NewAuthUseCase(users, blacklist, tokens, cfg.JWT.AccessTTLMinutes, cfg.JWT.RefreshTTLDays)

	httpHandler := httptransport.NewAuthHandler(authUC)
	echoSrv := httptransport.NewRouter(httpHandler, log)
	httpSrv := &http.Server{
		Addr:              net.JoinHostPort("", cfg.Server.Port),
		Handler:           echoSrv,
		ReadHeaderTimeout: 5 * time.Second,
	}

	grpcSrv := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcSrv, grpctransport.NewServer(authUC))

	grpcLis, err := net.Listen("tcp", net.JoinHostPort("", cfg.Server.GRPCPort))
	if err != nil {
		pool.Close()
		_ = rdb.Close()
		return nil, fmt.Errorf("grpc listen: %w", err)
	}

	return &App{
		cfg:     cfg,
		log:     log,
		pool:    pool,
		rdb:     rdb,
		httpSrv: httpSrv,
		grpcSrv: grpcSrv,
		grpcLis: grpcLis,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		a.log.Info("http server started", "port", a.cfg.Server.Port)
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http: %w", err)
		}
	}()

	go func() {
		a.log.Info("grpc server started", "port", a.cfg.Server.GRPCPort)
		if err := a.grpcSrv.Serve(a.grpcLis); err != nil {
			errCh <- fmt.Errorf("grpc: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	a.log.Info("shutting down")

	httpErr := a.httpSrv.Shutdown(ctx)

	stopped := make(chan struct{})
	go func() {
		a.grpcSrv.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-ctx.Done():
		a.grpcSrv.Stop()
	}

	_ = a.rdb.Close()
	a.pool.Close()
	return httpErr
}
