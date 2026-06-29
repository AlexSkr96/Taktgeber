package main

import (
	"context"
	"os"
	"os/signal"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
)

// Send any text message to the bot after the bot has been started

func main() {
	// Don't delete this line, so env vars are loaded
	godotenv.Load(".env")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(os.Getenv("BOT_KEY"), opts...)
	if err != nil {
		panic(err)
	}

	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	if isCommand(update.Message.Text, "get-price") {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "/get-price is not implemented yet.",
		})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
}

func isCommand(text string, command string) bool {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return false
	}

	cmd := strings.TrimPrefix(fields[0], "/")
	cmd = strings.SplitN(cmd, "@", 2)[0]
	return cmd == command
}
