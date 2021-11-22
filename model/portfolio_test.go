package model

import (
	"nuchal-api/util"
	"testing"
)

func TestGetPortfolio(t *testing.T) {

	userID := uint(1)

	p, err := GetPortfolio(userID)
	if err != nil {
		t.Fail()
	}

	util.PrettyPrint(p)
}
