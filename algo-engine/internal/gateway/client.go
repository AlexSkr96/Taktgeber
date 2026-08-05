package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexSkr96/Taktgeber/algo-engine/types"
	"github.com/gorilla/websocket"
)

var logger = log.Default()

const wsReadLimitBytes = 1024 * 1024

const getAccountStateEndpoint = "/api/v1/account/state"

const (
	reconnectBaseDelay = 1 * time.Second
	reconnectMaxDelay  = 30 * time.Second
)

// PriceWriter is the minimal interface the streamer needs from a store.
type PriceWriter interface {
	AddPrice(ctx context.Context, coin string, price float64, timestamp time.Time) error
}

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

func DecodeResponse(data []byte) (*types.SubResponse, error) {
	var resp types.SubResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) ReadNDecode(ctx context.Context) (*types.SubResponse, error) {
	data, err := c.Read(ctx)
	if err != nil {
		return nil, err
	}
	return DecodeResponse(data)
}

func DecodeAllMids(data []byte) (*types.AllMids, error) {
	var allMids types.AllMids
	if err := json.Unmarshal(data, &allMids); err != nil {
		return nil, fmt.Errorf("failed to unmarshal all mids: %w", err)
	}
	return &allMids, nil
}

func (c *Client) Subscribe(ctx context.Context, sub types.SubDetail) error {
	msg := types.Subscription{
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

// StreamPrices reads WebSocket messages and writes prices to the store.
// It handles reconnection with exponential backoff and exits when ctx is cancelled.
func (c *Client) StreamPrices(ctx context.Context, store PriceWriter, coins []string) {
	for {
		msg, err := c.ReadNDecode(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			logger.Printf("WebSocket read error: %v, reconnecting...", err)
			c.Close()
			if !c.reconnect(ctx) {
				return
			}
			continue
		}

		if msg.Channel != "allMids" {
			continue
		}

		allMids, err := DecodeAllMids(msg.Data)
		if err != nil {
			logger.Printf("Error decoding allMids: %v", err)
			continue
		}

		c.writePrices(ctx, store, coins, allMids)
	}
}

// reconnect attempts to reconnect and resubscribe with exponential backoff.
// Returns false if ctx is cancelled.
func (c *Client) reconnect(ctx context.Context) bool {
	delay := reconnectBaseDelay
	for {
		select {
		case <-ctx.Done():
			return false
		case <-time.After(delay):
		}

		logger.Printf("attempting to reconnect...")
		if err := c.Connect(ctx); err != nil {
			logger.Printf("reconnect failed: %v", err)
			delay = min(delay*2, reconnectMaxDelay)
			continue
		}
		if err := c.Subscribe(ctx, types.SubDetail{Type: "allMids"}); err != nil {
			logger.Printf("re-subscribe failed: %v", err)
			c.Close()
			delay = min(delay*2, reconnectMaxDelay)
			continue
		}
		logger.Println("Reconnected and re-subscribed!")
		return true
	}
}

func (c *Client) writePrices(ctx context.Context, store PriceWriter, coins []string, allMids *types.AllMids) {
	for _, coin := range coins {
		priceStr, ok := allMids.Mids[coin]
		if !ok {
			continue
		}

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			logger.Printf("Error while parsing coin price: %v", err)
			continue
		}

		if err := store.AddPrice(ctx, coin, price, time.Now().UTC()); err != nil {
			logger.Printf("Error while adding coin price: %v", err)
		}
	}
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
