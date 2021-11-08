package model

import (
	"fmt"
	"nuchal-api/db"
	"nuchal-api/util"
	"strings"
)

type Product struct {
	StrModel
	Name  string  `json:"name"`
	Base  string  `json:"base"`
	Quote string  `json:"quote"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Step  float64 `json:"step"`
}

func init() {
	db.Migrate(&Product{})
}

func FindAllProducts() []Product {
	var products []Product
	db.Resolve().Find(&products)
	return products
}

func (p Product) precise(f float64) string {
	decimal := util.FloatToDecimal(p.Step)
	zeros := len(strings.Split(decimal, ".")[1])
	zeroFormat := fmt.Sprintf(".%df", zeros)
	preciseFormat := "%" + zeroFormat
	return fmt.Sprintf(preciseFormat, f)
}
