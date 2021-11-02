package model

import (
	"fmt"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"gorm.io/gorm"
	"math"
	"nuchal-api/db"
	"nuchal-api/util"
	"strconv"
)

// Pattern defines the criteria for matching rates and placing orders.
type Pattern struct {
	gorm.Model

	// UserId
	UserID uint `json:"user_id"`

	// ProductID is concatenation of two currencies. e.g. BTC-USD
	ProductID uint `json:"product_id"`

	// Target is a percentage used to produce the goal sell price from the entry buy price.
	Target float64 `json:"target"`

	// Tolerance is a percentage used to derive a limit sell price from the entry buy price.
	Tolerance float64 `json:"tolerance"`

	// Size is the amount of the transaction, using the ProductMap native quote increment.
	Size float64 `json:"size"`

	// Delta is the size of an acceptable difference between tweezer bottom candlesticks.
	Delta float64 `json:"delta"`

	// Bind is represents the time of day when this strategy should activate.
	// Gross of bind is the total amount of milliseconds totaling hour and minutes.
	Bind int64 `json:"bind"`

	// Bound is the context to which this strategy looks to achieve so that it can break.
	// Values include buys, sells, holds, hours, minutes.
	Bound string `json:"bound"`

	// Break is a numerical value which gets applied to the Bound.
	Break int `json:"break"`

	// Enable is a flag that allows the system to bind, get bound, and break.
	Enable bool `json:"enable"`

	Product `json:"product" gorm:"embedded"`
}

func init() {
	db.Migrate(&Pattern{})
}

func (p *Pattern) GoalPrice(price float64) float64 {
	return price + (price * p.Target)
}

func (p *Pattern) LossPrice(price float64) float64 {
	return price - (price * p.Tolerance)
}

func (p *Pattern) MatchesTweezerBottomPattern(then, that, this Rate) bool {
	return then.IsInit() &&
		then.IsDown() &&
		that.IsInit() &&
		that.IsDown() &&
		this.IsUp() &&
		math.Abs(math.Min(that.Low, that.Close)-math.Min(this.Low, this.Open)) <= p.Delta
}

func (p Pattern) Save() Pattern {
	if p.ID > 0 {
		db.Resolve().Save(&p)
	} else {
		db.Resolve().Create(&p)
	}
	return p
}

func DeletePattern(patternID uint) {
	db.Resolve().Delete(&Pattern{}, patternID)
}

func FindPatternByID(patternID uint) Pattern {
	var pattern Pattern
	db.Resolve().
		Where("id = ?", pattern).
		Find(&pattern)
	return pattern
}

func GetPatterns(userID uint) []Pattern {
	var patterns []Pattern
	db.Resolve().
		Where("user_id = ?", userID).
		Find(&patterns)
	return patterns
}

func GetPattern(userID uint, productID uint) Pattern {
	var pattern Pattern
	db.Resolve().
		Where("user_id = ?", userID).
		Where("product_id = ?", productID).
		Find(&pattern)
	return pattern
}

func (p Pattern) NewMarketEntryOrder() cb.Order {
	return cb.Order{
		ProductID: strconv.Itoa(int(p.ProductID)),
		Side:      "buy",
		Size:      util.FloatToDecimal(p.Size),
		Type:      "market",
	}
}

func (p Pattern) NewMarketExitOrder() cb.Order {
	return cb.Order{
		ProductID: strconv.Itoa(int(p.ProductID)),
		Side:      "sell",
		Size:      util.FloatToDecimal(p.Size),
		Type:      "market",
	}
}

func (p Pattern) NewStopEntryOrder(size string, price float64) cb.Order {
	return cb.Order{
		Price:     precisePrice(p.ProductID, price),
		ProductID: strconv.Itoa(int(p.ProductID)),
		Side:      "sell",
		Size:      size,
		Type:      "limit",
		StopPrice: precisePrice(p.ProductID, price),
		Stop:      "entry",
	}
}

func (p Pattern) StopLossOrder(size string, price float64) cb.Order {
	return cb.Order{
		Price:     precisePrice(p.ProductID, price),
		ProductID: strconv.Itoa(int(p.ProductID)),
		Side:      "sell",
		Size:      size,
		Type:      "limit",
		StopPrice: precisePrice(p.ProductID, price),
		Stop:      "loss",
	}
}

//func FindPatterns(userID uint) []Pattern {
//
//	_ = InitProducts(userID)
//
//	var patterns []Pattern
//	for _, p := range GetPatterns(userID) {
//		pattern := Pattern{
//			Product: ProductMap[strconv.Itoa(int(p.ProductID))],
//		}
//		patterns = append(patterns, pattern)
//	}
//
//	return patterns
//}
//
//func FindPattern(userID uint, productID uint) Pattern {
//	p := GetPattern(userID, productID)
//	p.Product = ProductMap[strconv.Itoa(int(productID))]
//	return p
//}

func precisePrice(productID uint, price float64) string {
	product := ProductMap[productID]
	format := "%." + product.QuoteIncrement + "f"
	return fmt.Sprintf(format, price)
}
