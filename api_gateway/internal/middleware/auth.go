package middleware

import (
	"errors"
	"net/http"
	"strings"

	"api_gateway/pkg/jwtauth"

	"github.com/labstack/echo/v4"
)

const (
	HeaderUserID    = "X-User-Id"
	HeaderUserEmail = "X-User-Email"
	HeaderUsername  = "X-User-Username"
)

func JWTValidate(verifier *jwtauth.Verifier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if path == "/health" || path == "/metrics" || strings.HasPrefix(path, "/api/v1/auth") {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": jwtauth.ErrMissingHeader.Error()})
			}
			const prefix = "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": jwtauth.ErrInvalidFormat.Error()})
			}

			cl, err := verifier.VerifyAccess(strings.TrimPrefix(authHeader, prefix))
			if err != nil {
				msg := jwtauth.ErrInvalidToken.Error()
				if errors.Is(err, jwtauth.ErrWrongKind) {
					msg = jwtauth.ErrWrongKind.Error()
				}
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": msg})
			}

			c.Request().Header.Set(HeaderUserID, cl.UserID)
			c.Request().Header.Set(HeaderUserEmail, cl.Email)
			c.Request().Header.Set(HeaderUsername, cl.Username)
			return next(c)
		}
	}
}
