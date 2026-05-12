package gateway

import (
	"context"
	"encoding/json"
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

type Subscription struct {
	Method       string    `json:"method"`
	Subscription SubDetail `json:"subscription"`
}

type SubDetail struct {
	Type string `json:"type"`
	Coin string `json:"coin,omitempty"`
}

type SubResponse struct {
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

type AllMids struct {
	Mids map[string]string `json:"mids"`
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

func DecodeResponse(data []byte) (*SubResponse, error) {
	var resp SubResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) ReadNDecode(ctx context.Context) (*SubResponse, error) {
	data, err := c.Read(ctx)
	if err != nil {
		return nil, err
	}
	return DecodeResponse(data)
}

func DecodeAllMids(data []byte) (*AllMids, error) {
	var allMids AllMids
	if err := json.Unmarshal(data, &allMids); err != nil {
		return nil, fmt.Errorf("failed to unmarshal all mids: %w", err)
	}
	return &allMids, nil
}

func (c *Client) Subscribe(ctx context.Context, sub SubDetail) error {
	msg := Subscription{
		Method:       "subscribe",
		Subscription: sub,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription: %w", err)
	}

	if err := c.conn.Write(ctx, websocket.MessageText, data); err != nil {
		return fmt.Errorf("failed to send subscription: %w", err)
	}

	return nil
}
