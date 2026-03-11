package usecase

import (
	"auth_service/internal/domain"
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByID:       make(map[int64]*domain.User),
		usersByEmail:    make(map[string]*domain.User),
		usersByUsername: make(map[string]*domain.User),
		nextID:          0,
	}
}

type mockUserRepo struct {
	usersByID       map[int64]*domain.User
	usersByEmail    map[string]*domain.User
	usersByUsername map[string]*domain.User
	nextID          int64
	createErr       error
}

func (m *mockUserRepo) Create(ctx context.Context, email, username, passwordHash string) (*domain.User, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.nextID++
	u := &domain.User{
		ID:           m.nextID,
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
	}
	m.usersByID[u.ID] = u
	m.usersByEmail[email] = u
	m.usersByUsername[username] = u
	return u, nil
}

func (m *mockUserRepo) ByID(ctx context.Context, id int64) (*domain.User, error) {
	u, ok := m.usersByID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, ok := m.usersByEmail[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) ByUsername(ctx context.Context, username string) (*domain.User, error) {
	u, ok := m.usersByUsername[username]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func newMockBlacklist() *mockBlacklist {
	return &mockBlacklist{tokens: make(map[string]bool)}
}

type mockBlacklist struct {
	tokens map[string]bool
}

func (m *mockBlacklist) Add(ctx context.Context, tokenID string, ttlSeconds int) error {
	m.tokens[tokenID] = true
	return nil
}

func (m *mockBlacklist) Exists(ctx context.Context, tokenID string) (bool, error) {
	return m.tokens[tokenID], nil
}

type mockTokenProvider struct {
	access  string
	refresh string
	exp     time.Time
	parse   *domain.TokenClaims
	parseErr   error
	tokenID    string
	tokenIDErr error
}

func (m *mockTokenProvider) GenerateAccess(user *domain.User, ttlMinutes int) (string, time.Time, error) {
	return m.access, m.exp, nil
}

func (m *mockTokenProvider) GenerateRefresh(user *domain.User, ttlDays int) (string, time.Time, error) {
	return m.refresh, m.exp, nil
}

func (m *mockTokenProvider) ParseToken(tokenString string) (*domain.TokenClaims, error) {
	if m.parseErr != nil {
		return nil, m.parseErr
	}
	return m.parse, nil
}

func (m *mockTokenProvider) TokenID(tokenString string) (string, error) {
	if m.tokenIDErr != nil {
		return "", m.tokenIDErr
	}
	return m.tokenID, nil
}

func TestAuthUseCase_Register_Success(t *testing.T) {
	users := newMockUserRepo()
	bl := newMockBlacklist()
	tok := &mockTokenProvider{
		access:  "at",
		refresh: "rt",
		exp:     time.Now().Add(time.Hour),
	}
	uc := NewAuthUseCase(users, bl, tok, 15, 7)
	ctx := context.Background()

	user, pair, err := uc.Register(ctx, domain.RegisterInput{
		Email:    "a@b.c",
		Username: "user1",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if user == nil || user.Email != "a@b.c" || user.Username != "user1" {
		t.Errorf("unexpected user: %+v", user)
	}
	if pair == nil || pair.AccessToken != "at" || pair.RefreshToken != "rt" {
		t.Errorf("unexpected pair: %+v", pair)
	}
}

func TestAuthUseCase_Register_ShortPassword(t *testing.T) {
	uc := NewAuthUseCase(newMockUserRepo(), newMockBlacklist(), &mockTokenProvider{}, 15, 7)
	_, _, err := uc.Register(context.Background(), domain.RegisterInput{
		Email:    "a@b.c",
		Username: "u",
		Password: "short",
	})
	if err == nil {
		t.Fatal("expected error for short password")
	}
}

func TestAuthUseCase_Register_UserExists(t *testing.T) {
	users := newMockUserRepo()
	users.usersByEmail["a@b.c"] = &domain.User{ID: 1, Email: "a@b.c"}
	uc := NewAuthUseCase(users, newMockBlacklist(), &mockTokenProvider{}, 15, 7)
	_, _, err := uc.Register(context.Background(), domain.RegisterInput{
		Email:    "a@b.c",
		Username: "newuser",
		Password: "password123",
	})
	if !errors.Is(err, ErrUserExists) {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), BcryptCost)
	users := newMockUserRepo()
	users.usersByEmail["a@b.c"] = &domain.User{ID: 1, Email: "a@b.c", Username: "u1", PasswordHash: string(hash)}
	tok := &mockTokenProvider{access: "at", refresh: "rt", exp: time.Now().Add(time.Hour)}
	uc := NewAuthUseCase(users, newMockBlacklist(), tok, 15, 7)
	user, pair, err := uc.Login(context.Background(), domain.LoginInput{Login: "a@b.c", Password: "password123"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if user == nil || pair == nil || pair.AccessToken != "at" {
		t.Errorf("unexpected result: user=%+v pair=%+v", user, pair)
	}
}

func TestAuthUseCase_Login_InvalidCreds(t *testing.T) {
	uc := NewAuthUseCase(newMockUserRepo(), newMockBlacklist(), &mockTokenProvider{}, 15, 7)
	_, _, err := uc.Login(context.Background(), domain.LoginInput{
		Login:    "nobody@test.com",
		Password: "any",
	})
	if !errors.Is(err, ErrInvalidCreds) {
		t.Errorf("expected ErrInvalidCreds, got %v", err)
	}
}

func TestAuthUseCase_Refresh_Blacklisted(t *testing.T) {
	users := newMockUserRepo()
	users.usersByID[1] = &domain.User{ID: 1, Email: "a@b.c", Username: "u", PasswordHash: "hash"}
	bl := newMockBlacklist()
	bl.tokens["jti-123"] = true
	tok := &mockTokenProvider{
		parse:   &domain.TokenClaims{UserID: 1, Kind: domain.TokenKindRefresh, Exp: time.Now().Add(time.Hour), Iat: time.Now()},
		tokenID: "jti-123",
	}
	uc := NewAuthUseCase(users, bl, tok, 15, 7)
	_, err := uc.Refresh(context.Background(), "refresh-token")
	if !errors.Is(err, ErrTokenBlacklist) {
		t.Errorf("expected ErrTokenBlacklist, got %v", err)
	}
}

func TestAuthUseCase_Refresh_InvalidToken(t *testing.T) {
	tok := &mockTokenProvider{parseErr: errors.New("invalid")}
	uc := NewAuthUseCase(newMockUserRepo(), newMockBlacklist(), tok, 15, 7)
	_, err := uc.Refresh(context.Background(), "bad")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}
