package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"codeberg.org/a2100/Taktgeber/algo-engine/internal/gateway"
)

// apiConfig holds shared dependencies that HTTP handlers need access to.
type apiConfig struct {
	gatewayClient *gateway.Client
}

// Package variables
var logger = log.Default()
var client = gateway.NewClient(
	"ws://hl-gateway:8000/ws",
	"http://hl-gateway:8000",
)
var ctx = context.Background()

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// --- Existing gateway client setup ---
	client := gateway.NewClient(
		"ws://hl-gateway:8000/ws",
		"http://hl-gateway:8000",
	)

	// --- API server setup ---
	apiCfg := &apiConfig{
		gatewayClient: client,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", apiCfg.handlerHealth)
	mux.HandleFunc("GET /account", apiCfg.handlerAccount)

	srv := &http.Server{
		Addr:    ":9000",
		Handler: mux,
	}

	// Run the HTTP server in the background so it doesn't block
	// the WebSocket loop below.
	go func() {
		log.Println("API server listening on :9000")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// --- Your existing WebSocket / engine loop continues here ---
	// e.g. client.Connect(ctx), subscriptions, Redis writes, etc.

	<-ctx.Done()
	log.Println("shutting down")
}
