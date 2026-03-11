package http

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"runtime/debug"

	"github.com/labstack/echo/v4"
)

func PanicRecovery(log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if v := recover(); v != nil {
					log.Error("panic recovered", "panic", v, "stack", string(debug.Stack()))
					_ = c.JSON(500, map[string]string{"error": "internal server error"})
				}
			}()
			return next(c)
		}
	}
}

func RequestID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Request().Header.Get("X-Request-ID")
		if id == "" {
			id = generateRequestID()
		}
		c.Response().Header().Set("X-Request-ID", id)
		return next(c)
	}
}

func Logger(log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log.Info("request", "method", c.Request().Method, "path", c.Request().URL.Path, "remote", c.Request().RemoteAddr)
			return next(c)
		}
	}
}

func generateRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
