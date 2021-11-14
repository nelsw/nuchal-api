package model

import (
	"nuchal-api/util"
	"testing"
	"time"
)

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
	r := rates[len(rates)-1]
	util.PrettyPrint(r.Time())
}

func TestGetLastRate(t *testing.T) {
	var todayRate Rate

	FindFirstRateByProductIDInTimeDescOrder(productID, &todayRate)
	util.PrettyPrint(todayRate)

	before := time.Unix(todayRate.UnixSecond, 0).Add(time.Hour * -24).UTC().Unix()

	var yesterdayRate Rate
	FindFirstRateByProductIDAndLessThanTimeInTimeDescOrder(productID, before, &yesterdayRate)
	util.PrettyPrint(yesterdayRate)
}

func TestFindRates(t *testing.T) {

	//fmt.Println(alpha)
	//fmt.Println(omega)
	//
	//rates := FindRates(productID, alpha, omega)
	//
	//util.PrettyPrint(rates)
}
