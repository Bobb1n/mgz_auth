package domain

import "time"

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type LoginInput struct {
	Login    string
	Password string
}

type RegisterInput struct {
	Email    string
	Username string
	Password string
}
