package model

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
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
	productID = "1INCH-USD"
	omega     = time.Now().UTC().Unix()
	alpha     = time.Now().UTC().Add(time.Hour * 24 * 7 * -1).Unix()
)

func TestGetRates(t *testing.T) {

	rates, err := GetRates(userID, productID, alpha, omega)
	if err != nil {
		t.Fail()
	}
	util.PrettyPrint(rates)
}

func TestFindRates(t *testing.T) {

	//fmt.Println(alpha)
	//fmt.Println(omega)
	//
	//rates := FindRates(productID, alpha, omega)
	//
	//util.PrettyPrint(rates)
}
