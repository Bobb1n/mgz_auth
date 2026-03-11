package transport

import (
	"auth_service/internal/domain"
	"context"
)

type AuthUseCase interface {
	Register(ctx context.Context, in domain.RegisterInput) (*domain.User, *domain.TokenPair, error)
	Login(ctx context.Context, in domain.LoginInput) (*domain.User, *domain.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
}
