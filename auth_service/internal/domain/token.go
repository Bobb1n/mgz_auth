package domain

import "time"

const (
	TokenKindAccess  = "access"
	TokenKindRefresh = "refresh"
)

type TokenClaims struct {
	UserID   string
	Email    string
	Username string
	Kind     string
	Exp      time.Time
	Iat      time.Time
}
