package main

import (
	"context"
	"fmt"
	"log"

	"github.com/AlexSkr96/Taktgeber/algo-engine/internal/gateway"
)

func main() {
	fmt.Println("Taktgeber engine starting...")

	client := gateway.NewClient("ws://localhost:8000/ws")

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to gateway!")

	if err := client.Subscribe(ctx, gateway.SubDetail{Type: "allMids"}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Subscribed!")

	for {
		data, err := client.ReadNDecode(ctx)
		if err != nil {
			log.Fatal(err)
		}

		if data.Channel == "allMids" {
			fmt.Println("Channel:", data.Channel)
			// fmt.Println("Data:", string(data.Data))
			if allMids, err := gateway.DecodeAllMids(data.Data); err != nil {
				log.Fatal(err)
			} else {
				fmt.Println("Data:", allMids)
				fmt.Println("BTC:", allMids.Mids["BTC"])
			}
			break
		}
	}
}
