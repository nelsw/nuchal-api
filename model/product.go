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

var ProductIDs []uint
var ProductArr []Product
var ProductMap = map[uint]Product{}
var usdRegex = regexp.MustCompile(`^((\w{3,5})(-USD))$`)

func init() {
	db.Migrate(&Product{})
}

func InitProducts(userID uint) error {

	u := FindUserByID(userID)

	db.Resolve().Find(&ProductArr)

	if len(ProductArr) > 0 {
		for _, product := range ProductArr {
			ProductMap[product.Model.ID] = product
			ProductIDs = append(ProductIDs, product.Model.ID)
		}
		sort.SliceStable(ProductIDs, func(i, j int) bool {
			return ProductIDs[i] < ProductIDs[j]
		})
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
		ProductMap[p.Model.ID] = p
		ProductIDs = append(ProductIDs, p.Model.ID)
		ProductArr = append(ProductArr, p)
	}

	return nil

}
