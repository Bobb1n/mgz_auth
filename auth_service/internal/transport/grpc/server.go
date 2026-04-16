package grpctransport

import (
	"context"
	"errors"

	"auth_service/internal/domain"
	"auth_service/internal/usecase"
	authv1 "auth_service/pkg/api/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type authUseCase interface {
	Register(ctx context.Context, in domain.RegisterInput) (*domain.User, *domain.TokenPair, error)
	Login(ctx context.Context, in domain.LoginInput) (*domain.User, *domain.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
}

type Server struct {
	authv1.UnimplementedAuthServiceServer
	uc authUseCase
}

func NewServer(uc authUseCase) *Server {
	return &Server{uc: uc}
}

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthResponse, error) {
	user, pair, err := s.uc.Register(ctx, domain.RegisterInput{
		Email:    req.GetEmail(),
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.AuthResponse{
		User:   toProtoUser(user),
		Tokens: toProtoTokenPair(pair),
	}, nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error) {
	user, pair, err := s.uc.Login(ctx, domain.LoginInput{
		Login:    req.GetLogin(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.AuthResponse{
		User:   toProtoUser(user),
		Tokens: toProtoTokenPair(pair),
	}, nil
}

func (s *Server) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.TokenPair, error) {
	pair, err := s.uc.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoTokenPair(pair), nil
}

func (s *Server) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := s.uc.Logout(ctx, req.GetAccessToken(), req.GetRefreshToken()); err != nil {
		return nil, toGRPCError(err)
	}
	return &authv1.LogoutResponse{}, nil
}

func toProtoUser(u *domain.User) *authv1.User {
	if u == nil {
		return nil
	}
	return &authv1.User{
		Id:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		CreatedAt: timestamppb.New(u.CreatedAt),
	}
}

func toProtoTokenPair(p *domain.TokenPair) *authv1.TokenPair {
	if p == nil {
		return nil
	}
	return &authv1.TokenPair{
		AccessToken:  p.AccessToken,
		RefreshToken: p.RefreshToken,
		ExpiresAt:    timestamppb.New(p.ExpiresAt),
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrUserExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, usecase.ErrInvalidCreds),
		errors.Is(err, usecase.ErrInvalidToken),
		errors.Is(err, usecase.ErrTokenBlacklist):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, usecase.ErrShortPassword):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		if isValidationMsg(err.Error()) {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		return status.Error(codes.Internal, "internal error")
	}
}

func isValidationMsg(msg string) bool {
	switch msg {
	case "email is required", "email is invalid", "username is required":
		return true
	}
	return false
}
