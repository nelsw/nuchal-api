package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"gorm.io/gorm"
	"nuchal-api/db"
	"regexp"
)

type Product struct {
	gorm.Model
	cb.Product
}

var ProductIDs []string
var ProductArr []cb.Product
var ProductMap = map[string]cb.Product{}
var usdRegex = regexp.MustCompile(`^((\w{3,5})(-USD))$`)

func init() {
	db.Migrate(&Product{})
}

func InitProducts(userID uint) error {

	u := FindUserByID(userID)

	allProducts, err := u.Client().GetProducts()
	if err != nil {
		return err
	}

	for _, product := range allProducts {
		if product.BaseCurrency == "DAI" ||
			product.BaseCurrency == "USDT" ||
			product.BaseMinSize == "" ||
			product.QuoteIncrement == "" ||
			!usdRegex.MatchString(product.ID) {
			continue
		}
		ProductMap[product.ID] = product
		ProductArr = append(ProductArr, product)
		ProductIDs = append(ProductIDs, product.ID)
	}

	return nil

}
