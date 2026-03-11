package domain

import "time"

const (
	TokenKindAccess  = "access"
	TokenKindRefresh = "refresh"
)

type TokenClaims struct {
	UserID   int64     `json:"user_id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	Kind     string    `json:"kind"`
	Exp      time.Time `json:"exp"`
	Iat      time.Time `json:"iat"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type LoginInput struct {
	Login    string `json:"login"`    // email или username
	Password string `json:"password"`
}

type RegisterInput struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}
