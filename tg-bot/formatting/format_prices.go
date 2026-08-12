package formatting

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/AlexSkr96/Taktgeber/algo-engine/types"
)

func FormatPrices(pricePoints []types.PricePoint) string {
	log.Printf("PricePoint slice for formatting: %v\n", pricePoints)

	var b strings.Builder

	for _, pricePoint := range pricePoints {
		displayTime := time.UnixMilli(pricePoint.UnixTimestamp).Format("02.01 15:04:05 '06")
		// displayTime := pricePoint.UnixTimestamp
		fmt.Fprintf(&b, "%v: $%v\n", displayTime, pricePoint.Price)
	}

	return b.String()
}
