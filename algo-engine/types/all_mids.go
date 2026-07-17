package types

import "encoding/json"

// Structs for coin info

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
