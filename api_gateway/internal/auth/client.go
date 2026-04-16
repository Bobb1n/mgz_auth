package auth

import (
	"fmt"

	authv1 "auth_service/pkg/api/auth/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	cc  *grpc.ClientConn
	api authv1.AuthServiceClient
}

func NewClient(addr string) (*Client, error) {
	cc, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, fmt.Errorf("dial auth grpc: %w", err)
	}
	return &Client{
		cc:  cc,
		api: authv1.NewAuthServiceClient(cc),
	}, nil
}

func (c *Client) Close() error {
	if c == nil || c.cc == nil {
		return nil
	}
	return c.cc.Close()
}
