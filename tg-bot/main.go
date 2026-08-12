package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AlexSkr96/Taktgeber/tg-bot/commands"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_KEY"))
	if err != nil {
		log.Panicf("Failed to connect to TG bot API: %v\n", err)
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

		// Extract potential args
		args := strings.Fields(update.Message.CommandArguments())

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "health":
			msg.Text, err = commands.Health(args)
		case "account":
			msg.Text, err = commands.Account(args)
		case "price":
			msg.Text, err = commands.Price(args)
		default:
			msg.Text = "I don't know that command"
		}

		if err != nil {
			msg.Text = fmt.Sprintf("Failed to execute command: %v", err)
		}

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Failed to send message: %v\nMesasge content: %v\n", err, msg.Text)
		}
	}
}
