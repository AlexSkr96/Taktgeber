package commands

import (
	"fmt"
	"net/http"
)

const healthHelp = "Verifies connection to backend and it's status."

func Health(args []string) (string, error) {
	if isHelp(args) {
		return healthHelp, nil
	}

	resp, err := http.Get(healthURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return fmt.Sprintf("Algo-engine status: %s", resp.Status), nil
}
