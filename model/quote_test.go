package model

import (
	"nuchal-api/util"
	"testing"
)

func TestGetQuotes(t *testing.T) {

	quotes := GetQuotes(uint(1))
	util.PrettyPrint(quotes)

}
