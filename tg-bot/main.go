package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"codeberg.org/a2100/Taktgeber/algo-engine/types"
	"codeberg.org/a2100/Taktgeber/tg-bot/formatting"
)

const algoEngineURL = "http://algo-engine:9000"
const healthURL = algoEngineURL + "/health"
const accountURL = algoEngineURL + "/account"

func main() {
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

	switch update.Message.Text {
	case "/health":
		healthHandler(ctx, b, update)
	case "/account":
		accountHandler(ctx, b, update)
	}
}

func healthHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	resp, err := http.Get(healthURL)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Failed to connect: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Algo-engine status: %s", resp.Status),
	})
}

func accountHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	accountState := types.AccountState{}

	resp, err := http.Get(accountURL)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Filed to get account data: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&accountState); err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Failed to decode account data: %v", err),
		})
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      formatting.FormatAccountState(accountState),
		ParseMode: models.ParseModeHTML,
	})
}
