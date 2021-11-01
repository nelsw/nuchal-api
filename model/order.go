package model

import (
	"fmt"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"gorm.io/gorm"
	"nuchal-api/db"
	"nuchal-api/util"
)

func init() {
	db.Migrate(&Order{})
}

type Order struct {
	gorm.Model
	Order cb.Order `gorm:"embedded;embeddedPrefix:cb_"`
}

func SaveOrder(order cb.Order) {
	db.Resolve().Create(&Order{gorm.Model{}, order})
}

func NewMarketEntryOrder(productID string, size float64) cb.Order {
	return cb.Order{
		ProductID: productID,
		Side:      "buy",
		Size:      util.FloatToDecimal(size),
		Type:      "market",
	}
}

func NewStopEntryOrder(productID, size string, price float64) cb.Order {
	return cb.Order{
		Price:     precisePrice(productID, price),
		ProductID: productID,
		Side:      "sell",
		Size:      size,
		Type:      "limit",
		StopPrice: precisePrice(productID, price),
		Stop:      "entry",
	}
}

func StopLossOrder(productID, size string, price float64) cb.Order {
	return cb.Order{
		Price:     precisePrice(productID, price),
		ProductID: productID,
		Side:      "sell",
		Size:      size,
		Type:      "limit",
		StopPrice: precisePrice(productID, price),
		Stop:      "loss",
	}
}

func precisePrice(productID string, price float64) string {
	product := ProductMap[productID]
	format := "%." + product.QuoteIncrement + "f"
	return fmt.Sprintf(format, price)
}
