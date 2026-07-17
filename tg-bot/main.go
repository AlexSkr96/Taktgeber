package main

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	algoEngineURL = "http://algo-engine:9000"
	healthURL     = algoEngineURL + "/health"
	accountURL    = algoEngineURL + "/account"
	priceURL      = algoEngineURL + "/price"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		// Create a new MessageConfig. We don't have text yet,
		// so we leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "health":
			msg.Text, err = commandHealth()
		case "account":
			msg.Text, err = commandAccount()
		case "price":
			msg.Text, err = commandPrice()
		default:
			msg.Text = "I don't know that command"
		}

		if err != nil {
			msg.Text = fmt.Sprint("Failed to execute command: %v", err)
		}

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message: %v\nMesasge content: %v\n", err, msg.Text)
		}
	}
}
