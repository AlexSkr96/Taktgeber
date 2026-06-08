package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
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
		msg, err := client.ReadNDecode(ctx)
		if err != nil {
			log.Fatal(err)
		}

		// fmt.Println("Raw Data:", msg)
		// fmt.Println("Channel:", msg.Channel)
		// fmt.Println("Data:", string(msg.Data))
		if msg.Channel == "allMids" {
			if allMids, err := gateway.DecodeAllMids(msg.Data); err != nil {
				log.Fatal(err)
			} else {
				coins := []string{"BTC", "ETH", "XMR"}
				// fmt.Println("Data:", allMids)
				// fmt.Println(time.Now().UTC(), "BTC:", allMids.Mids["BTC"])
				// fmt.Println(time.Now().UTC(), "ETH:", allMids.Mids["ETH"])
				// fmt.Println(time.Now().UTC(), "XMR:", allMids.Mids["XMR"])

				for _, coin := range coins {
					timestamp := time.Now().UTC()
					fmt.Println(timestamp, coin+":", allMids.Mids[coin])

					price, err := strconv.ParseFloat(allMids.Mids[coin], 64)
					if err != nil {
						fmt.Println(fmt.Errorf("Error while parsing coin price: ", err))
					}

					store.AddPricePoint(ctx, coin, price, timestamp)
				}
			}
			// break
		}
	}
}
