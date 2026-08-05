package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AlexSkr96/Taktgeber/algo-engine/types"
	"github.com/AlexSkr96/Taktgeber/tg-bot/formatting"
)

func commandAccount() (string, error) {
	accountState := types.AccountState{}

	resp, err := http.Get(accountURL)
	if err != nil {
		return "", fmt.Errorf("Failed to get account state: %v", err)
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&accountState); err != nil {
		return "", fmt.Errorf("Failed to decode account data: %v", err)
	}

	return formatting.FormatAccountState(accountState), nil
}
