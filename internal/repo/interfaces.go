package repo

import (
	"auth_service/internal/domain"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, email, username, passwordHash string) (*domain.User, error)
	ByID(ctx context.Context, id int64) (*domain.User, error)
	ByEmail(ctx context.Context, email string) (*domain.User, error)
	ByUsername(ctx context.Context, username string) (*domain.User, error)
}

type BlacklistRepository interface {
	Add(ctx context.Context, tokenID string, ttlSeconds int) error
	Exists(ctx context.Context, tokenID string) (bool, error)
}
