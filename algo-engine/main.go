package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/AlexSkr96/Taktgeber/algo-engine/internal/gateway"
	"github.com/AlexSkr96/Taktgeber/algo-engine/internal/store"
)

func main() {
	logger := log.Default()

	logger.Println("Taktgeber engine starting...")

	// client := gateway.NewClient("ws://localhost:8000/ws")
	client := gateway.NewClient(
		"ws://hl-gateway:8000/ws",
		"http://hl-gateway:8000",
	)

	ctx := context.Background()

	// Connect with retry - gateway might not be ready yet
	reconnectBaseDelay := 1 * time.Second
	reconnectMaxDelay := 30 * time.Second
	delay := reconnectBaseDelay
	for {
		if err := client.Connect(ctx); err != nil {
			logger.Printf("failed to connect to gateway: %v, retrying in %v...", err, delay)
			time.Sleep(delay)
			delay = min(delay*2, reconnectMaxDelay)
			continue
		}
		break
	}

	fmt.Println("Connected to gateway!")

	// Subscribe with retry
	delay = reconnectBaseDelay
	for {
		if err := client.Subscribe(ctx, gateway.SubDetail{Type: "allMids"}); err != nil {
			logger.Printf("failed to subscribe: %v, retrying in %v...", err, delay)
			time.Sleep(delay)
			delay = min(delay*2, reconnectMaxDelay)
			continue
		}
		break
	}

	logger.Println("Subscribed!")

	// Redis
	store := store.NewStore("redis:6379")

	if err := store.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	logger.Println("Connected to Redis!")

	// Channel processing
	coins := []string{"BTC", "ETH", "XMR"}

	//Get recent prices from Redis
	go func() {
		for {
			time.Sleep(30 * time.Second)
			for _, coin := range coins {
				pricePoints, err := store.GetRecentPrices(ctx, coin, 60*time.Hour)
				if err != nil {
					logger.Println(fmt.Errorf("ERROR while getting prices from redis: %v", err))
				}
				logger.Printf("Recent %v prices: %v\n", coin, pricePoints)
			}
		}
	}()

	//Get current prices from Hyperliquid and write them to Redis
	go func() {
		for {
			msg, err := client.ReadNDecode(ctx)
			if err != nil {
				logger.Printf("WebSocket read error: %v, reconnecting...", err)
				client.Close()

				delay := reconnectBaseDelay
				for {
					time.Sleep(delay)
					logger.Printf("attempting to reconnect...")
					if err := client.Connect(ctx); err != nil {
						logger.Printf("reconnect failed: %v", err)
						delay = min(delay*2, reconnectMaxDelay)
						continue
					}
					if err := client.Subscribe(ctx, gateway.SubDetail{Type: "allMids"}); err != nil {
						logger.Printf("re-subscribe failed: %v", err)
						client.Close()
						delay = min(delay*2, reconnectMaxDelay)
						continue
					}
					logger.Println("Reconnected and re-subscribed!")
					break
				}
				continue
			}

			// fmt.Println("Raw Data:", msg)
			// fmt.Println("Channel:", msg.Channel)
			// fmt.Println("Data:", string(msg.Data))
			if msg.Channel == "allMids" {
				if allMids, err := gateway.DecodeAllMids(msg.Data); err != nil {
					log.Fatal(err)
				} else {
					// fmt.Println("Data:", allMids)
					// fmt.Println(time.Now().UTC(), "BTC:", allMids.Mids["BTC"])
					// fmt.Println(time.Now().UTC(), "ETH:", allMids.Mids["ETH"])
					// fmt.Println(time.Now().UTC(), "XMR:", allMids.Mids["XMR"])

					for _, coin := range coins {
						timestamp := time.Now().UTC()
						// fmt.Println(timestamp, coin+":", allMids.Mids[coin])

						price, err := strconv.ParseFloat(allMids.Mids[coin], 64)
						if err != nil {
							logger.Println(fmt.Errorf("Error while parsing coin price: %v", err))
						}

						err = store.AddPrice(ctx, coin, price, timestamp)
						if err != nil {
							logger.Println(fmt.Errorf("Error while adding coin price: %v", err))
						}
					}
				}
			}
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Printf("received signal %v, shutting down...", sig)
	client.Close()
}
