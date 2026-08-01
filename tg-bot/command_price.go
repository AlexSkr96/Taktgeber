package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"codeberg.org/a2100/Taktgeber/algo-engine/types"
	"codeberg.org/a2100/Taktgeber/tg-bot/formatting"
)

func commandPrice(coin, since, limit string) (string, error) {
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

	log.Printf("commandPrice; resp.Body: %v\n", resp.Body)
	if err = json.NewDecoder(resp.Body).Decode(&pricePoints); err != nil {
		return "", fmt.Errorf("Failed to decode prices: %v", err)
	}

	return formatting.FormatPrices(pricePoints), nil
}
