package model

import (
	"nuchal-api/util"
	"testing"
)

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
