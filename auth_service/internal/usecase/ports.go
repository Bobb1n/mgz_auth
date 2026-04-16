package usecase

import (
	"context"
	"time"

	"auth_service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, email, username, passwordHash string) (*domain.User, error)
	ByID(ctx context.Context, id string) (*domain.User, error)
	ByEmail(ctx context.Context, email string) (*domain.User, error)
	ByUsername(ctx context.Context, username string) (*domain.User, error)
}

type BlacklistRepository interface {
	Add(ctx context.Context, tokenID string, ttlSeconds int) error
	Exists(ctx context.Context, tokenID string) (bool, error)
}

type TokenProvider interface {
	GenerateAccess(user *domain.User, ttlMinutes int) (string, time.Time, error)
	GenerateRefresh(user *domain.User, ttlDays int) (string, time.Time, error)
	ParseToken(tokenString string) (*domain.TokenClaims, error)
	TokenID(tokenString string) (string, error)
}
