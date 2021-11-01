package service

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

var (
	userID    = uint(1)
	productID = "ALGO-USD"
	alpha     = time.Date(2021, 10, 31, 0, 0, 0, 0, time.UTC).UnixMilli()
	omega     = time.Date(2021, 10, 31, 11, 0, 0, 0, time.UTC).UnixMilli()
)

func TestSaveAllRatesBetweenFor(t *testing.T) {

	if err := SaveTodayRatesFor(userID, productID); err != nil {
		t.Fail()
	}

}

func TestSaveAllNewRates(t *testing.T) {
	if err := SaveAllNewRates(userID); err != nil {
		t.Fail()
	}
}
