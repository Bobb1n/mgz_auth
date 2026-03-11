package internal

import (
	"auth_service/internal/domain"
	"time"
)

// TokenProvider — контракт для usecase (удобно мокать в тестах).
type TokenProvider interface {
	GenerateAccess(user *domain.User, ttlMinutes int) (string, time.Time, error)
	GenerateRefresh(user *domain.User, ttlDays int) (string, time.Time, error)
	ParseToken(tokenString string) (*domain.TokenClaims, error)
	TokenID(tokenString string) (string, error)
}
