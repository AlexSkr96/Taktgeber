package types

import "encoding/json"

type AccountState struct {
	Status string             `json:"status"`
	Result AccountStateResult `json:"result"`
}

type AccountStateResult struct {
	AccountAddress string      `json:"account_address"`
	UserState      UserState   `json:"user_state"`
	OpenOrders     []OpenOrder `json:"open_orders"`
}

type UserState struct {
	MarginSummary              MarginSummary   `json:"marginSummary"`
	CrossMarginSummary         MarginSummary   `json:"crossMarginSummary"`
	CrossMaintenanceMarginUsed json.Number     `json:"crossMaintenanceMarginUsed"`
	Withdrawable               json.Number     `json:"withdrawable"`
	AssetPositions             []AssetPosition `json:"assetPositions"`
	Time                       int64           `json:"time"`
}

type MarginSummary struct {
	AccountValue    json.Number `json:"accountValue"`
	TotalNtlPos     json.Number `json:"totalNtlPos"`
	TotalRawUsd     json.Number `json:"totalRawUsd"`
	TotalMarginUsed json.Number `json:"totalMarginUsed"`
}

type AssetPosition struct {
	Type     string   `json:"type"`
	Position Position `json:"position"`
}

type Position struct {
	Coin           string      `json:"coin"`
	EntryPx        json.Number `json:"entryPx"`
	Szi            json.Number `json:"szi"`
	PositionValue  json.Number `json:"positionValue"`
	UnrealizedPnl  json.Number `json:"unrealizedPnl"`
	ReturnOnEquity json.Number `json:"returnOnEquity"`
	MarginUsed     json.Number `json:"marginUsed"`
	LiquidationPx  json.Number `json:"liquidationPx"`
	MaxLeverage    json.Number `json:"maxLeverage"`
}

type OpenOrder struct {
	Coin      string      `json:"coin"`
	LimitPx   json.Number `json:"limitPx"`
	Oid       json.Number `json:"oid"`
	OrigSz    json.Number `json:"origSz"`
	Side      string      `json:"side"`
	Sz        json.Number `json:"sz"`
	Timestamp json.Number `json:"timestamp"`
}
