package usecase

import (
	"auth_service/internal/domain"
	"auth_service/internal"
	"auth_service/internal/repo"
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists    = errors.New("user with such email or username already exists")
	ErrInvalidCreds  = errors.New("invalid login or password")
	ErrInvalidToken  = errors.New("invalid or expired token")
	ErrTokenBlacklist = errors.New("token has been revoked")
)

const (
	MinPasswordLen = 8
	BcryptCost     = 10
)

type AuthUseCase struct {
	users            repo.UserRepository
	blacklist        repo.BlacklistRepository
	tokens           internal.TokenProvider
	accessTTLMinutes int
	refreshTTLDays   int
}

func NewAuthUseCase(
	users repo.UserRepository,
	blacklist repo.BlacklistRepository,
	tokens internal.TokenProvider,
	accessTTLMinutes, refreshTTLDays int,
) *AuthUseCase {
	return &AuthUseCase{
		users:            users,
		blacklist:        blacklist,
		tokens:           tokens,
		accessTTLMinutes: accessTTLMinutes,
		refreshTTLDays:   refreshTTLDays,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, in domain.RegisterInput) (*domain.User, *domain.TokenPair, error) {
	if len(in.Password) < MinPasswordLen {
		return nil, nil, fmt.Errorf("password must be at least %d characters", MinPasswordLen)
	}
	_, err := uc.users.ByEmail(ctx, in.Email)
	if err == nil {
		return nil, nil, ErrUserExists
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, nil, err
	}
	_, err = uc.users.ByUsername(ctx, in.Username)
	if err == nil {
		return nil, nil, ErrUserExists
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), BcryptCost)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}
	user, err := uc.users.Create(ctx, in.Email, in.Username, string(hash))
	if err != nil {
		return nil, nil, err
	}
	pair, err := uc.issueTokenPair(user)
	if err != nil {
		return user, nil, err
	}
	return user, pair, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, in domain.LoginInput) (*domain.User, *domain.TokenPair, error) {
	var user *domain.User
	var err error
	if isEmail(in.Login) {
		user, err = uc.users.ByEmail(ctx, in.Login)
	} else {
		user, err = uc.users.ByUsername(ctx, in.Login)
	}
	if err != nil || user == nil {
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return nil, nil, err
		}
		return nil, nil, ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return nil, nil, ErrInvalidCreds
	}
	pair, err := uc.issueTokenPair(user)
	if err != nil {
		return nil, nil, err
	}
	return user, pair, nil
}

func (uc *AuthUseCase) Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	claims, err := uc.tokens.ParseToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if claims.Kind != domain.TokenKindRefresh {
		return nil, ErrInvalidToken
	}
	tokenID, err := uc.tokens.TokenID(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}
	ok, err := uc.blacklist.Exists(ctx, tokenID)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, ErrTokenBlacklist
	}
	user, err := uc.users.ByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return uc.issueTokenPair(user)
}

func (uc *AuthUseCase) Logout(ctx context.Context, accessToken, refreshToken string) error {
	// Blacklist both tokens until their natural expiry
	for _, raw := range []string{accessToken, refreshToken} {
		if raw == "" {
			continue
		}
		claims, err := uc.tokens.ParseToken(raw)
		if err != nil {
			continue // ignore invalid, still try to blacklist by id
		}
		tokenID, err := uc.tokens.TokenID(raw)
		if err != nil {
			continue
		}
		ttl := int(claims.Exp.Unix() - claims.Iat.Unix())
		if ttl <= 0 {
			ttl = 1
		}
		_ = uc.blacklist.Add(ctx, tokenID, ttl)
	}
	return nil
}

func (uc *AuthUseCase) issueTokenPair(user *domain.User) (*domain.TokenPair, error) {
	access, expAccess, err := uc.tokens.GenerateAccess(user, uc.accessTTLMinutes)
	if err != nil {
		return nil, err
	}
	refresh, expRefresh, err := uc.tokens.GenerateRefresh(user, uc.refreshTTLDays)
	if err != nil {
		return nil, err
	}
	exp := expAccess
	if expRefresh.After(exp) {
		exp = expRefresh
	}
	return &domain.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    exp,
	}, nil
}

func isEmail(s string) bool {
	return contains(s, '@')
}

func contains(s string, c rune) bool {
	for _, r := range s {
		if r == c {
			return true
		}
	}
	return false
}
