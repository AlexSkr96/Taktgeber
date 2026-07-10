package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"codeberg.org/a2100/Taktgeber/algo-engine/types"
	"github.com/gorilla/websocket"
)

var logger = log.Default()

const wsReadLimitBytes = 1024 * 1024

const getAccountStateEndpoint = "/api/v1/account/state"

type Client struct {
	wsUrl   string
	httpUrl string
	conn    *websocket.Conn
	http    *http.Client
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func NewClient(wsUrl string, httpUrl string) *Client {
	return &Client{
		wsUrl:   wsUrl,
		httpUrl: httpUrl,
		http:    &http.Client{},
	}
}

func (c *Client) Connect(ctx context.Context) error {
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, c.wsUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}
	conn.SetReadLimit(wsReadLimitBytes)
	c.conn = conn
	return nil
}

func (c *Client) Read(ctx context.Context) ([]byte, error) {
	_, data, err := c.conn.ReadMessage()
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

	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send subscription: %w", err)
	}

	return nil
}

func (c *Client) GetAccountState(ctx context.Context) (types.AccountState, error) {
	accountState := types.AccountState{}
	url := fmt.Sprintf("%s%s", c.httpUrl, getAccountStateEndpoint)

	resp, err := c.http.Get(url)
	if err != nil {
		return accountState, fmt.Errorf("error while getting account state: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return accountState, fmt.Errorf("error while reading account state body: %w", err)
	}

	logger.Printf("Account state body: %v\n", string(body))

	if err = json.Unmarshal(body, &accountState); err != nil {
		return accountState, fmt.Errorf("error while unmarshalling account state: %w", err)
	}

	return accountState, nil
}
