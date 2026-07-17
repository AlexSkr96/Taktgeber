package main

import (
	"fmt"
	"net/http"
	"strconv"

	"codeberg.org/a2100/Taktgeber/algo-engine/types"
)

func commandPrice(coin, since, limit string) ([]types.PricePoint, error) {
	pricePoints := []types.PricePoint{}

	thisPriceURL := priceURL
	if coin != "" {
		thisPriceURL += fmt.Sprintf("?coin=%v", coin)
	} else {
		return "", fmt.Errorf("Coin param can't be null")
	}

	if since != "" {
		thisPriceURL += fmt.Sprintf("&since=%v", since)
	}

	if limit != "" {
		if _, err := strconv.ParseInt(limit, 10, 64); err != nil {
			return "", fmt.Errorf("Invalid limit %v, must be an integer number", limit)
		}
		thisPriceURL += fmt.Sprintf("&limit=%v", limit)
	}

	resp, err := http.Get(thisPriceURL)
	if err != nil {
		return "", fmt.Errorf("Failed to get prices from redis: %v", err)
	}
	defer resp.Body.Close()

	// if err = json.NewDecoder(resp.Body).Decode()

	return "Not yet implemented", err
}
