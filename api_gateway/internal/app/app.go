package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"api_gateway/internal/auth"
	"api_gateway/internal/chat"
	"api_gateway/internal/config"
	"api_gateway/internal/middleware"
	"api_gateway/internal/swipe"
	"api_gateway/internal/user"
	"api_gateway/pkg/jwtauth"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type App struct {
	cfg     *config.Config
	log     *slog.Logger
	echo    *echo.Echo
	httpSrv *http.Server

	authClient  *auth.Client
	chatClient  *chat.Client
	userClient  *user.Client
	swipeClient *swipe.Client
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	authClient, err := auth.NewClient(cfg.AuthGRPCAddr)
	if err != nil {
		return nil, fmt.Errorf("auth grpc client: %w", err)
	}
	chatClient, err := chat.NewClient(cfg.ChatGRPCAddr)
	if err != nil {
		_ = authClient.Close()
		return nil, fmt.Errorf("chat grpc client: %w", err)
	}
	userClient, err := user.NewClient(cfg.UserGRPCAddr)
	if err != nil {
		_ = authClient.Close()
		_ = chatClient.Close()
		return nil, fmt.Errorf("user grpc client: %w", err)
	}
	swipeClient, err := swipe.NewClient(cfg.SwipeGRPCAddr)
	if err != nil {
		_ = authClient.Close()
		_ = chatClient.Close()
		_ = userClient.Close()
		return nil, fmt.Errorf("swipe grpc client: %w", err)
	}

	verifier := jwtauth.NewVerifier([]byte(cfg.JWTSecret))

	e := echo.New()
	e.HideBanner = true

	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.CORS())

	// Публичные роуты
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "ok",
			"service": "api-gateway",
		})
	})

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Защищённые роуты
	api := e.Group("")
	api.Use(middleware.JWTValidate(verifier))

	auth.NewHandler(authClient).RegisterRoutes(api)
	chat.NewHandler(chatClient).RegisterRoutes(api)
	user.NewHandler(userClient).RegisterRoutes(api)
	swipe.NewHandler(swipeClient).RegisterRoutes(api)

	httpSrv := &http.Server{
		Addr:              net.JoinHostPort("", cfg.ServerPort),
		Handler:           e,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		cfg:         cfg,
		log:         log,
		echo:        e,
		httpSrv:     httpSrv,
		authClient:  authClient,
		chatClient:  chatClient,
		userClient:  userClient,
		swipeClient: swipeClient,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		a.log.Info("gateway starting",
			"port", a.cfg.ServerPort,
			"auth_grpc", a.cfg.AuthGRPCAddr,
			"chat_grpc", a.cfg.ChatGRPCAddr,
			"user_grpc", a.cfg.UserGRPCAddr,
			"swipe_grpc", a.cfg.SwipeGRPCAddr,
		)

		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http: %w", err)
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
	a.log.Info("shutting down gateway")

	err := a.httpSrv.Shutdown(ctx)

	_ = a.authClient.Close()
	_ = a.chatClient.Close()
	_ = a.userClient.Close()
	_ = a.swipeClient.Close()

	return err
}
