package chat

import (
	"context"
	"fmt"
	"time"

	chatv1 "gitlab.com/siffka/chat-message-mgz/pkg/api/chat/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	cc  *grpc.ClientConn
	api chatv1.ChatMessageServiceClient
}

func NewClient(addr string) (*Client, error) {
	cc, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)
	if err != nil {
		return nil, fmt.Errorf("dial chat grpc: %w", err)
	}
	return &Client{
		cc:  cc,
		api: chatv1.NewChatMessageServiceClient(cc),
	}, nil
}

func (c *Client) Close() error {
	if c == nil || c.cc == nil {
		return nil
	}
	return c.cc.Close()
}

func (c *Client) CreateDirectChat(ctx context.Context, user1ID, user2ID string) (*chatv1.Chat, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.api.CreateDirectChat(ctx, &chatv1.CreateDirectChatRequest{
		User1Id: user1ID,
		User2Id: user2ID,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetChat(), nil
}

func (c *Client) SendMessage(ctx context.Context, chatID, senderID, text string) (*chatv1.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.api.SendMessage(ctx, &chatv1.SendMessageRequest{
		ChatId:       chatID,
		SenderUserId: senderID,
		Text:         text,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetMessage(), nil
}

func (c *Client) ListUserChats(ctx context.Context, userID string, limit, offset int32) ([]*chatv1.ChatPreview, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.api.ListUserChats(ctx, &chatv1.ListUserChatsRequest{
		UserId: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetChats(), nil
}

