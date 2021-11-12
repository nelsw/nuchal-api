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

func TestFindAllProducts(t *testing.T) {

	products, err := FindAllProducts()
	if err != nil {
		t.Fail()
	}
	util.PrettyPrint(&products)

}

func TestFindAllProductsByQuote(t *testing.T) {

	products, err := FindAllProductsByQuote("USD")
	if err != nil {
		t.Fail()
	}
	util.PrettyPrint(products)

}

func TestFindProductByID(t *testing.T) {

	p, err := FindProductByID("ALGO-USD")
	if err != nil {
		t.Fail()
	}
	util.PrettyPrint(p)

}
