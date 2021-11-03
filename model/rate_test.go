package model

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
	omega     = time.Now().Unix()
	alpha     = time.Now().Add(time.Hour * 24 * 7 * -1).Unix()
)

func TestFindRatesBetween(t *testing.T) {
	rates := FindRatesBetween(productID, alpha, omega)
	fmt.Println(len(rates))
}

//func TestInitRate(t *testing.T) {
//	if err := InitRate(userID, "ALGO-USD"); err != nil {
//		fmt.Println(err)
//		t.Fail()
//	}
//}
//
func TestInitRates(t *testing.T) {
	if err := InitRates(userID); err != nil {
		fmt.Println(err)
		t.Fail()
	}
}
