package model

import (
	"nuchal-api/util"
	"testing"
)

func TestNewProductChart(t *testing.T) {

	chart, err := NewProductChart(userID, productID, alpha, omega)
	if err != nil {
		t.Fail()
	}

	util.PrettyPrint(chart)
}
