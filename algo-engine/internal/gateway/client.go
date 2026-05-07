package gateway

import (
	"context"
	"fmt"

	"github.com/coder/websocket"
)

type Client struct {
	url  string
	conn *websocket.Conn
}

func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	conn, _, err := websocket.Dial(ctx, c.url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}

	c.conn = conn
	return nil
}

func (c *Client) Read(ctx context.Context) ([]byte, error) {
	_, data, err := c.conn.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read from gateway: %w", err)
	}

	return data, nil
}
