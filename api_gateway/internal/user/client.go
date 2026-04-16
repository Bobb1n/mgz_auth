package user

import (
	"fmt"

	userv1 "github.com/S1FFFkA/user-mgz/pkg/api/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	cc  *grpc.ClientConn
	api userv1.UserServiceClient
}

func NewClient(addr string) (*Client, error) {
	cc, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, fmt.Errorf("dial user grpc: %w", err)
	}
	return &Client{
		cc:  cc,
		api: userv1.NewUserServiceClient(cc),
	}, nil
}

func (c *Client) Close() error {
	if c == nil || c.cc == nil {
		return nil
	}
	return c.cc.Close()
}
