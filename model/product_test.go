package model

import (
	"fmt"
	"nuchal-api/util"
	"testing"
)

func TestInitProducts(t *testing.T) {

	userID := 1
	err := InitProducts(uint(userID))
	if err != nil {
		t.Fail()
	}

	fmt.Println(util.Pretty(ProductArr))

}