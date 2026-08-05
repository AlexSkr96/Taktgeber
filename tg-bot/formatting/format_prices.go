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
		displayTime := time.Unix(pricePoint.UnixTimestamp, 0).Format("01.02 15:04:05")
		b.WriteString(fmt.Sprintf("%v: $%v", displayTime, pricePoint.Price))
	}

	return b.String()
}
