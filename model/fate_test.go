package model

import (
	"testing"
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
