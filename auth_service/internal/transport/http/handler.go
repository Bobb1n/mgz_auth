package http

import (
	"context"
	"errors"
	"net/http"

	"auth_service/internal/domain"
	"auth_service/internal/usecase"

	"github.com/labstack/echo/v4"
)

type AuthService interface {
	Register(ctx context.Context, in domain.RegisterInput) (*domain.User, *domain.TokenPair, error)
	Login(ctx context.Context, in domain.LoginInput) (*domain.User, *domain.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
}

type AuthHandler struct {
	uc AuthService
}

func NewAuthHandler(uc AuthService) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid json"})
	}

	user, pair, err := h.uc.Register(c.Request().Context(), domain.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return h.writeErr(c, err)
	}
	return c.JSON(http.StatusCreated, authResponse{
		User:   newUserResponse(user),
		Tokens: newTokenPairResponse(pair),
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid json"})
	}

	user, pair, err := h.uc.Login(c.Request().Context(), domain.LoginInput{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		return h.writeErr(c, err)
	}
	return c.JSON(http.StatusOK, authResponse{
		User:   newUserResponse(user),
		Tokens: newTokenPairResponse(pair),
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req refreshRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid json"})
	}
	if req.RefreshToken == "" {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "refresh_token is required"})
	}

	pair, err := h.uc.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return h.writeErr(c, err)
	}
	return c.JSON(http.StatusOK, newTokenPairResponse(pair))
}

func (h *AuthHandler) Logout(c echo.Context) error {
	var req logoutRequest
	_ = c.Bind(&req)
	if err := h.uc.Logout(c.Request().Context(), req.AccessToken, req.RefreshToken); err != nil {
		return h.writeErr(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) writeErr(c echo.Context, err error) error {
	switch {
	case errors.Is(err, usecase.ErrUserExists):
		return c.JSON(http.StatusConflict, errorResponse{Error: err.Error()})
	case errors.Is(err, usecase.ErrInvalidCreds),
		errors.Is(err, usecase.ErrInvalidToken),
		errors.Is(err, usecase.ErrTokenBlacklist):
		return c.JSON(http.StatusUnauthorized, errorResponse{Error: err.Error()})
	case errors.Is(err, usecase.ErrShortPassword):
		return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:
		if isValidationErr(err) {
			return c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func isValidationErr(err error) bool {
	switch err.Error() {
	case "email is required", "email is invalid", "username is required":
		return true
	}
	return false
}
