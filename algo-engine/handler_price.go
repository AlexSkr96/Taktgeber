package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func (cfg *apiConfig) handlerPrice(w http.ResponseWriter, r *http.Request) {
	coin := r.URL.Query().Get("coin")
	if coin == "" {
		http.Error(w, "missing 'coin' query parameter", http.StatusBadRequest)
		return
	}

	// Optional 'since' parameter (default 24h), e.g. "1h", "30m", "168h"
	since := 24 * time.Hour
	if s := r.URL.Query().Get("since"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid 'since' value: %v", err), http.StatusBadRequest)
			return
		}
		since = d
	}

	// Optional 'limit' parameter (default 100)
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		n, err := strconv.Atoi(l)
		if err != nil || n <= 0 {
			http.Error(w, "invalid 'limit' value", http.StatusBadRequest)
			return
		}
		limit = n
	}

	prices, err := cfg.store.GetRecentPrices(r.Context(), coin, since)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Apply limit
	if len(prices) > limit {
		prices = prices[len(prices)-limit:]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prices)
}
