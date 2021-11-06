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
	productID = uint(1)
	omega     = time.Now().Unix()
	alpha     = time.Now().Add(time.Hour * 24 * 7 * -1).Unix()
)

func TestGetAllRatesBetween(t *testing.T) {

	fmt.Println(time.Unix(1636082173, 0))
	fmt.Println(time.Unix(1633403773, 0))

	rates := GetAllRatesBetween(userID, productID, 1633403773, 1636082173)
	util.PrettyPrint(rates)

}
