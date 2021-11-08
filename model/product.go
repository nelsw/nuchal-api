package model

import (
	"fmt"
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

type Crypto struct {
	gorm.Model
	Symbol string  `json:"symbol"`
	Quote  string  `json:"quote"`
	Min    float64 `json:"min"`
	Max    int     `json:"max"`
	Step   float64 `json:"step"`
}

type Currency struct {
	gorm.Model
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

var ProductIDs []uint
var ProductArr []Product
var ProductMap = map[uint]Product{}
var usdRegex = regexp.MustCompile(`^((\w{3,5})(-USD))$`)

func init() {
	db.Migrate(&Currency{})
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

func FindAllProducts(userID uint) []Product {
	err := InitProducts(userID)
	if err != nil {
		fmt.Println(err)
	}
	var products []Product
	db.Resolve().Find(&products)
	return products
}

func findPID(productID uint) string {
	var product Product
	db.Resolve().
		Where("id = ?", productID).
		Find(&product)
	return product.BaseCurrency + "-" + product.QuoteCurrency
}

func (p Product) precise(f float64) string {
	return fmt.Sprintf("%"+fmt.Sprintf(".%df", len(strings.Split(p.QuoteIncrement, ".")[1])), f)
}
