package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// claims совпадают со структурой токена auth_service (access).
type claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Kind     string `json:"kind"`
	jwt.RegisteredClaims
}

const (
	HeaderUserID    = "X-User-Id"
	HeaderUserEmail = "X-User-Email"
	HeaderUsername  = "X-User-Username"
)

// JWTValidate проверяет Bearer-токен и кладёт user_id, email, username в заголовки запроса.
// Не вызывается для /health и /api/v1/auth — там токен не нужен.
func JWTValidate(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if path == "/health" || strings.HasPrefix(path, "/api/v1/auth") {
				return next(c)
			}

			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
			}
			const prefix = "Bearer "
			if !strings.HasPrefix(auth, prefix) {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization format"})
			}
			tokenStr := strings.TrimPrefix(auth, prefix)

			tok, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(*jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			}
			cl, ok := tok.Claims.(*claims)
			if !ok || !tok.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			}
			if cl.Kind != "access" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "access token required"})
			}

			c.Request().Header.Set(HeaderUserID, cl.UserID)
			c.Request().Header.Set(HeaderUserEmail, cl.Email)
			c.Request().Header.Set(HeaderUsername, cl.Username)
			return next(c)
		}
	}
}
