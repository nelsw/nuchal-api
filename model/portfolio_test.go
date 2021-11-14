package model

import (
	"testing"
)

func TestGetPortfolio(t *testing.T) {

	userID := uint(1)

	_, err := GetPortfolio(userID)
	if err != nil {
		t.Fail()
	}
}

func TestSellFills(t *testing.T) {

	_ = SellFills(uint(1))

}
