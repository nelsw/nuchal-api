package view

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"testing"
	"time"
)

func init() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("***%s****", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%s", i))
	}
}

func Test(t *testing.T) {

	for _, game := range FindAllGames() {
		for _, play := range game.Plays {
			fmt.Println(fmt.Sprintf(headFmt, "ID", "Open", "High", "Low", "Close"))

			for _, rate := range play.Rates {
				fmt.Println(fmt.Sprintf(lineFmt, rate.Stamp(), rate.Open, rate.High, rate.Low, rate.Close))
			}

			fmt.Println(game.ProductID)
			fmt.Println(play.Bonus)
			fmt.Println(play.Ratio)
			fmt.Println()
			break
		}
	}
}
