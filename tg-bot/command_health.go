package main

import (
	"fmt"
	"net/http"
)

func commandHealth() (string, error) {
	resp, err := http.Get(healthURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return fmt.Sprintf("Algo-engine status: %s", resp.Status), nil
}
