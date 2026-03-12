package http

import (
	"log/slog"

	"github.com/labstack/echo/v4"
)

func NewRouter(auth *AuthHandler, log *slog.Logger) *echo.Echo {
	e := echo.New()

	e.Use(PanicRecovery(log))
	e.Use(RequestID)
	e.Use(Logger(log))

	g := e.Group("/api/v1/auth")
	g.POST("/register", auth.Register)
	g.POST("/login", auth.Login)
	g.POST("/refresh", auth.Refresh)
	g.POST("/logout", auth.Logout)

	return e
}
