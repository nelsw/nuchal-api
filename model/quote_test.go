package model

import (
	"nuchal-api/util"
	"testing"
)

func TestGetQuotes(t *testing.T) {

	quotes, err := GetQuotes()
	if err != nil {
		t.Fail()
	}
	util.PrettyPrint(quotes)

}
