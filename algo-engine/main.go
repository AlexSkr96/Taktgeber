package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AlexSkr96/Taktgeber/algo-engine/internal/gateway"
	"github.com/AlexSkr96/Taktgeber/algo-engine/internal/store"
)

func main() {
	fmt.Println("Taktgeber engine starting...")

	// client := gateway.NewClient("ws://localhost:8000/ws")
	client := gateway.NewClient("ws://hl-gateway:8000/ws")

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to gateway!")

	if err := client.Subscribe(ctx, gateway.SubDetail{Type: "allMids"}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Subscribed!")

	// Redis
	store := store.NewStore("redis:6379")

	if err := store.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to Redis!")

	// Channel processing
	for {
		data, err := client.ReadNDecode(ctx)
		if err != nil {
			log.Fatal(err)
		}

		if data.Channel == "allMids" {
			// fmt.Println("Channel:", data.Channel)
			// fmt.Println("Data:", string(data.Data))
			if allMids, err := gateway.DecodeAllMids(data.Data); err != nil {
				log.Fatal(err)
			} else {
				// fmt.Println("Data:", allMids)
				fmt.Println(time.Now().UTC(), "BTC:", allMids.Mids["BTC"])
				fmt.Println(time.Now().UTC(), "ETH:", allMids.Mids["ETH"])
			}
			// break
		}
	}
}
