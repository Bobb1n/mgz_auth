package http

import (
	"auth_service/internal/domain"
	"auth_service/internal/transport"
	"auth_service/internal/usecase/logic"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	uc transport.AuthUseCase
}

func NewAuthHandler(uc transport.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var in domain.RegisterInput
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	user, pair, err := h.uc.Register(c.Request().Context(), in)
	if err != nil {
		return h.writeErr(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{
		"user":          userResponse(user),
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_at":    pair.ExpiresAt,
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var in domain.LoginInput
	if err := c.Bind(&in); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	user, pair, err := h.uc.Login(c.Request().Context(), in)
	if err != nil {
		return h.writeErr(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"user":          userResponse(user),
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_at":    pair.ExpiresAt,
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
	}
	if body.RefreshToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "refresh_token required"})
	}
	pair, err := h.uc.Refresh(c.Request().Context(), body.RefreshToken)
	if err != nil {
		return h.writeErr(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_at":    pair.ExpiresAt,
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	var body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.Bind(&body)
	if err := h.uc.Logout(c.Request().Context(), body.AccessToken, body.RefreshToken); err != nil {
		return h.writeErr(c, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) writeErr(c echo.Context, err error) error {
	switch {
	case errors.Is(err, usecase.ErrUserExists):
		return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, usecase.ErrInvalidCreds):
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	case errors.Is(err, usecase.ErrInvalidToken), errors.Is(err, usecase.ErrTokenBlacklist):
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func userResponse(u *domain.User) map[string]any {
	if u == nil {
		return nil
	}
	return map[string]any{
		"id":         u.ID,
		"email":      u.Email,
		"username":   u.Username,
		"created_at": u.CreatedAt,
	}
}
