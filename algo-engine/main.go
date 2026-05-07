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

	for {
		data, err := client.Read(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Received:", string(data))
	}
}
