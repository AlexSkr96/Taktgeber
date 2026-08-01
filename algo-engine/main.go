package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"codeberg.org/a2100/Taktgeber/algo-engine/internal/gateway"
	"codeberg.org/a2100/Taktgeber/algo-engine/internal/store"
	"codeberg.org/a2100/Taktgeber/algo-engine/types"
)

// apiConfig holds shared dependencies that HTTP handlers need access to.
type apiConfig struct {
	gatewayClient *gateway.Client
	store         store.Store
}

// Package variables
var logger = log.Default()
var client = gateway.NewClient(
	"ws://hl-gateway:8000/ws",
	"http://hl-gateway:8000",
)
var ctx = context.Background()
var port = ":9000"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// --- API server setup ---
	apiCfg := &apiConfig{
		gatewayClient: client,
		store:         *store.NewStore("redis:6379"),
	}

	if err := apiCfg.store.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	// Connect with retry
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	// Subscribe
	if err := client.Subscribe(ctx, types.SubDetail{Type: "allMids"}); err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", apiCfg.handlerHealth)
	mux.HandleFunc("GET /account", apiCfg.handlerAccount)
	mux.HandleFunc("GET /price", apiCfg.handlerPrice)

	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	// Run the HTTP server in the background
	go func() {
		log.Printf("API server listening on %v\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Stream prices to Redis (handles reconnection internally)
	coins := []string{"BTC", "ETH", "XMR"}
	go client.StreamPrices(ctx, &apiCfg.store, coins)

	<-ctx.Done()
	log.Println("shutting down")
	client.Close()
	time.Sleep(100 * time.Millisecond)
}
