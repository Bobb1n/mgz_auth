package jwtauth

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrMissingHeader = errors.New("missing authorization header")
	ErrInvalidFormat = errors.New("invalid authorization format")
	ErrInvalidToken  = errors.New("invalid or expired token")
	ErrWrongKind     = errors.New("access token required")
)

const KindAccess = "access"

type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Kind     string `json:"kind"`
	jwt.RegisteredClaims
}

type Verifier struct {
	secret []byte
}

func NewVerifier(secret []byte) *Verifier {
	return &Verifier{secret: secret}
}

func (v *Verifier) Verify(token string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(token, &Claims{}, func(*jwt.Token) (interface{}, error) {
		return v.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	cl, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, ErrInvalidToken
	}
	return cl, nil
}

func (v *Verifier) VerifyAccess(token string) (*Claims, error) {
	cl, err := v.Verify(token)
	if err != nil {
		return nil, err
	}
	if cl.Kind != KindAccess {
		return nil, ErrWrongKind
	}
	return cl, nil
}
