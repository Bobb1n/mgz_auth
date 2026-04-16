package usecase

import (
	"errors"
	"fmt"
	"time"

	"auth_service/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

type jwtClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Kind     string `json:"kind"`
	JTI      string `json:"jti"`
	jwt.RegisteredClaims
}

type JWTProvider struct {
	secret []byte
}

func NewJWTProvider(secret string) *JWTProvider {
	return &JWTProvider{secret: []byte(secret)}
}

func (p *JWTProvider) GenerateAccess(user *domain.User, ttlMinutes int) (string, time.Time, error) {
	exp := time.Now().Add(time.Duration(ttlMinutes) * time.Minute)
	return p.generate(user, domain.TokenKindAccess, exp)
}

func (p *JWTProvider) GenerateRefresh(user *domain.User, ttlDays int) (string, time.Time, error) {
	exp := time.Now().Add(time.Duration(ttlDays) * 24 * time.Hour)
	return p.generate(user, domain.TokenKindRefresh, exp)
}

func (p *JWTProvider) generate(user *domain.User, kind string, exp time.Time) (string, time.Time, error) {
	jti := fmt.Sprintf("%s-%s-%d", user.ID, kind, time.Now().UnixNano())
	c := jwtClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Kind:     kind,
		JTI:      jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := tok.SignedString(p.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

func (p *JWTProvider) ParseToken(tokenString string) (*domain.TokenClaims, error) {
	tok, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(*jwt.Token) (interface{}, error) {
		return p.secret, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := tok.Claims.(*jwtClaims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return &domain.TokenClaims{
		UserID:   c.UserID,
		Email:    c.Email,
		Username: c.Username,
		Kind:     c.Kind,
		Exp:      c.ExpiresAt.Time,
		Iat:      c.IssuedAt.Time,
	}, nil
}

func (p *JWTProvider) TokenID(tokenString string) (string, error) {
	tok, _, err := jwt.NewParser().ParseUnverified(tokenString, &jwtClaims{})
	if err != nil {
		return "", err
	}
	c, ok := tok.Claims.(*jwtClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	return c.JTI, nil
}
