package swipe

import (
	"fmt"

	swipev1 "swipe-mgz/pkg/api/swipe/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	cc  *grpc.ClientConn
	api swipev1.SwipeServiceClient
}

func NewClient(addr string) (*Client, error) {
	cc, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, fmt.Errorf("dial swipe grpc: %w", err)
	}
	return &Client{
		cc:  cc,
		api: swipev1.NewSwipeServiceClient(cc),
	}, nil
}

func (c *Client) Close() error {
	if c == nil || c.cc == nil {
		return nil
	}
	return c.cc.Close()
}
