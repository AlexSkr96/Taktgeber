package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handlerAccount(w http.ResponseWriter, r *http.Request) {
	accountState, err := client.GetAccountState(ctx)
	if err != nil {
		error := fmt.Errorf("Failed to get account state: %v", err)
		logger.Println(error)
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(accountState); err != nil {
		error := fmt.Errorf("Failed to encode account state: %v", err)
		logger.Println(error)
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return
	}
}
