package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"auth_service/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists     = errors.New("user with such email or username already exists")
	ErrInvalidCreds   = errors.New("invalid login or password")
	ErrInvalidToken   = errors.New("invalid or expired token")
	ErrTokenBlacklist = errors.New("token has been revoked")
	ErrShortPassword  = fmt.Errorf("password must be at least %d characters", MinPasswordLen)
)

const (
	MinPasswordLen = 8
	BcryptCost     = 10
)

type AuthUseCase struct {
	users            UserRepository
	blacklist        BlacklistRepository
	tokens           TokenProvider
	accessTTLMinutes int
	refreshTTLDays   int
}

func NewAuthUseCase(
	users UserRepository,
	blacklist BlacklistRepository,
	tokens TokenProvider,
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
	if err := validateRegister(in); err != nil {
		return nil, nil, err
	}

	if exists, err := uc.userExists(ctx, in.Email, in.Username); err != nil {
		return nil, nil, err
	} else if exists {
		return nil, nil, ErrUserExists
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
	user, err := uc.findByLogin(ctx, in.Login)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil, ErrInvalidCreds
		}
		return nil, nil, err
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
	if err != nil || claims.Kind != domain.TokenKindRefresh {
		return nil, ErrInvalidToken
	}

	tokenID, err := uc.tokens.TokenID(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	revoked, err := uc.blacklist.Exists(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("check blacklist: %w", err)
	}
	if revoked {
		return nil, ErrTokenBlacklist
	}

	user, err := uc.users.ByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return uc.issueTokenPair(user)
}

func (uc *AuthUseCase) Logout(ctx context.Context, accessToken, refreshToken string) error {
	for _, raw := range []string{accessToken, refreshToken} {
		if raw == "" {
			continue
		}
		claims, err := uc.tokens.ParseToken(raw)
		if err != nil {
			continue
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

func (uc *AuthUseCase) userExists(ctx context.Context, email, username string) (bool, error) {
	if _, err := uc.users.ByEmail(ctx, email); err == nil {
		return true, nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return false, err
	}
	if _, err := uc.users.ByUsername(ctx, username); err == nil {
		return true, nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return false, err
	}
	return false, nil
}

func (uc *AuthUseCase) findByLogin(ctx context.Context, login string) (*domain.User, error) {
	if strings.Contains(login, "@") {
		return uc.users.ByEmail(ctx, login)
	}
	return uc.users.ByUsername(ctx, login)
}

func (uc *AuthUseCase) issueTokenPair(user *domain.User) (*domain.TokenPair, error) {
	access, expAccess, err := uc.tokens.GenerateAccess(user, uc.accessTTLMinutes)
	if err != nil {
		return nil, fmt.Errorf("generate access: %w", err)
	}
	refresh, expRefresh, err := uc.tokens.GenerateRefresh(user, uc.refreshTTLDays)
	if err != nil {
		return nil, fmt.Errorf("generate refresh: %w", err)
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

func validateRegister(in domain.RegisterInput) error {
	if strings.TrimSpace(in.Email) == "" {
		return errors.New("email is required")
	}
	if !strings.Contains(in.Email, "@") {
		return errors.New("email is invalid")
	}
	if strings.TrimSpace(in.Username) == "" {
		return errors.New("username is required")
	}
	if len(in.Password) < MinPasswordLen {
		return ErrShortPassword
	}
	return nil
}
