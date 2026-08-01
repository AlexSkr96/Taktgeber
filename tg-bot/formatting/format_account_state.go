package formatting

import (
	"fmt"
	"log"
	"strings"

	"codeberg.org/a2100/Taktgeber/algo-engine/types"
)

func FormatAccountState(state types.AccountState) string {
	log.Printf("AccountState for formatting: %+v\n", state)

	var b strings.Builder

	ms := state.Result.UserState.MarginSummary
	accountValue, _ := ms.AccountValue.Float64()
	marginUsed, _ := ms.TotalMarginUsed.Float64()
	withdrawable, _ := state.Result.UserState.Withdrawable.Float64()

	b.WriteString("💰 Account Summary\n")
	b.WriteString(fmt.Sprintf("Equity: $%.2f\n", accountValue))
	b.WriteString(fmt.Sprintf("Margin used: $%.2f\n", marginUsed))
	b.WriteString(fmt.Sprintf("Withdrawable: $%.2f\n\n", withdrawable))

	positions := state.Result.UserState.AssetPositions
	if len(positions) == 0 {
		b.WriteString("📊 Positions\nNone open\n\n")
	} else {
		b.WriteString("📊 Positions\n")
		for _, ap := range positions {
			p := ap.Position
			size, _ := p.Szi.Float64()
			entry, _ := p.EntryPx.Float64()
			pnl, _ := p.UnrealizedPnl.Float64()
			roe, _ := p.ReturnOnEquity.Float64()

			side := "LONG"
			if size < 0 {
				side = "SHORT"
			}

			emoji := "🟢"
			if pnl < 0 {
				emoji = "🔴"
			}

			b.WriteString(fmt.Sprintf(
				"%s %s %s %.4f @ %.2f\n   PnL: $%.2f (%.2f%%)\n",
				emoji, p.Coin, side, size, entry, pnl, roe*100,
			))
		}
		b.WriteString("\n")
	}

	orders := state.Result.OpenOrders
	if len(orders) == 0 {
		b.WriteString("📋 Open Orders\nNone\n")
	} else {
		b.WriteString("📋 Open Orders\n")
		for _, o := range orders {
			px, _ := o.LimitPx.Float64()
			sz, _ := o.Sz.Float64()

			side := "BUY"
			emoji := "🟢"
			if o.Side == "A" {
				side = "SELL"
				emoji = "🔴"
			}

			b.WriteString(fmt.Sprintf("%s %s %s %.4f @ %.2f\n", emoji, o.Coin, side, sz, px))
		}
	}

	return b.String()
}
