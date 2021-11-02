package model

import (
	"fmt"
	"nuchal-api/util"
	"testing"
)

func TestGetPortfolio(t *testing.T) {

	userID := uint(1)

	portfolio, err := GetPortfolio(userID)
	if err != nil {
		t.Fail()
	}

	fmt.Println(util.Pretty(portfolio))
}
