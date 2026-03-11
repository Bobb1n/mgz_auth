package usecase

import (
	"auth_service/internal/domain"
	"testing"
	"time"
)

func TestJWTProvider_GenerateAndParse(t *testing.T) {
	p := NewJWTProvider("test-secret")
	user := &domain.User{ID: 1, Email: "a@b.c", Username: "u1"}

	access, exp, err := p.GenerateAccess(user, 15)
	if err != nil {
		t.Fatalf("GenerateAccess: %v", err)
	}
	if access == "" {
		t.Fatal("empty access token")
	}
	if exp.Before(time.Now()) {
		t.Error("exp already in past")
	}

	claims, err := p.ParseToken(access)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != 1 || claims.Email != "a@b.c" || claims.Kind != domain.TokenKindAccess {
		t.Errorf("unexpected claims: %+v", claims)
	}

	jti, err := p.TokenID(access)
	if err != nil {
		t.Fatalf("TokenID: %v", err)
	}
	if jti == "" {
		t.Error("empty jti")
	}
}

func TestJWTProvider_ParseInvalidToken(t *testing.T) {
	p := NewJWTProvider("secret")
	_, err := p.ParseToken("invalid.jwt.here")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
