package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/AlexSkr96/Taktgeber/algo-engine/types"
	"github.com/AlexSkr96/Taktgeber/tg-bot/formatting"
)

var priceHelp = "Shows list of recent prices for a given coin.\n" +
	"Usage: /price {coin} {since} {limit}. {since} and {limit} are optional, 24h and 50 are default values.\n" +
	"Example: /price BTC 24h 50"

func Price(args []string) (string, error) {
	if isHelp(args) {
		return priceHelp, nil
	}

	coin := args[0]
	since := "24h"
	if len(args) > 1 {
		since = args[1]
	}
	limit := "50"
	if len(args) > 2 {
		limit = args[2]
	}

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
