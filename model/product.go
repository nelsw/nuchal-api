package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"gorm.io/gorm"
	"nuchal-api/db"
	"regexp"
	"sort"
	"strings"
)

type Product struct {
	gorm.Model
	cb.Product
}

var ProductIDs []string
var ProductArr []Product
var ProductMap = map[string]Product{}
var usdRegex = regexp.MustCompile(`^((\w{3,5})(-USD))$`)

func init() {
	db.Migrate(&Product{})
}

func InitProducts(userID uint) error {

	u := FindUserByID(userID)

	db.Resolve().Find(&ProductArr)

	if len(ProductArr) > 0 {
		for _, product := range ProductArr {
			ProductMap[product.Product.ID] = product
			ProductIDs = append(ProductIDs, product.Product.ID)
		}
		return nil
	}

	products, err := u.Client().GetProducts()
	if err != nil {
		return err
	}

	sort.SliceStable(products, func(i, j int) bool {
		return strings.Compare(products[i].ID, products[j].ID) < 0
	})

	for _, product := range products {
		if product.BaseCurrency == "DAI" ||
			product.BaseCurrency == "USDT" ||
			product.BaseMinSize == "" ||
			product.QuoteIncrement == "" ||
			!usdRegex.MatchString(product.ID) {
			continue
		}
		p := Product{gorm.Model{}, product}
		db.Resolve().Create(&p)
		ProductMap[p.Product.ID] = p
		ProductIDs = append(ProductIDs, p.Product.ID)
		ProductArr = append(ProductArr, p)
	}

	return nil

}
