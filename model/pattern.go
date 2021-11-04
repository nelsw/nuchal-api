package model

import (
	"fmt"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"math"
	"nuchal-api/db"
	"nuchal-api/util"
	"strings"
)

// Pattern defines the criteria for matching rates and placing orders.
type Pattern struct {
	gorm.Model

	// UserId
	UserID uint `json:"user_id"`

	// Currency is concatenation of two currencies. e.g. BTC-USD
	ProductID uint `json:"product_id" gorm:"product_id"`

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

	Product Product `json:"product"`
}

func init() {
	db.Migrate(&Pattern{})
}

func (p Pattern) Logger() *zerolog.Logger {
	logger := log.
		With().
		Uint("userID", p.UserID).
		Uint("patternID", p.Model.ID).
		Str("productID", p.Currency()).
		Logger()
	return &logger
}

func (p Pattern) Currency() string {
	return p.Product.BaseCurrency + "-" + p.Product.QuoteCurrency
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

func (p *Pattern) Save() {
	if p.Model.ID > 0 {
		db.Resolve().Save(p)
	} else {
		db.Resolve().Create(p)
	}
}

func (p Pattern) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Uint("userID", p.UserID).
		Uint("patternID", p.Model.ID).
		Str("productID", p.Currency())
}

func DeletePattern(patternID uint) {
	db.Resolve().Delete(&Pattern{}, patternID)
}

func FindPatternByID(patternID uint) Pattern {
	var pattern Pattern
	db.Resolve().
		Preload("Product").
		Where("id = ?", patternID).
		Find(&pattern)
	return pattern
}

func GetPatterns(userID uint) []Pattern {
	var patterns []Pattern
	db.Resolve().
		Preload("Product").
		Where("user_id = ?", userID).
		Find(&patterns)
	return patterns
}

func FindPattern(id uint) Pattern {
	var pattern Pattern
	db.Resolve().
		Preload("Product").
		Where("id = ?", id).
		Find(&pattern)
	return pattern
}

func GetPattern(userID uint, productID uint) Pattern {
	var pattern Pattern
	db.Resolve().
		Preload("Product").
		Where("user_id = ?", userID).
		Where("product_id = ?", productID).
		Find(&pattern)
	return pattern
}

func (p Pattern) NewMarketEntryOrder() cb.Order {
	return cb.Order{
		ProductID: p.Currency(),
		Side:      "buy",
		Size:      util.FloatToDecimal(p.Size),
		Type:      "market",
	}
}

func (p Pattern) NewMarketExitOrder() cb.Order {
	return cb.Order{
		ProductID: p.Currency(),
		Side:      "sell",
		Size:      util.FloatToDecimal(p.Size),
		Type:      "market",
	}
}

func (p Pattern) NewStopEntryOrder(size string, price float64) cb.Order {
	return cb.Order{
		Price:     p.precisePrice(price),
		ProductID: p.Currency(),
		Side:      "sell",
		Size:      size,
		Type:      "limit",
		StopPrice: p.precisePrice(price),
		Stop:      "entry",
	}
}

func (p Pattern) StopLossOrder(size string, price float64) cb.Order {
	return cb.Order{
		Price:     p.precisePrice(price),
		ProductID: p.Currency(),
		Side:      "sell",
		Size:      size,
		Type:      "limit",
		StopPrice: p.precisePrice(price),
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
//			Product: ProductMap[strconv.Itoa(int(p.Currency))],
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

func (p Pattern) precisePrice(price float64) string {
	return fmt.Sprintf("%"+fmt.Sprintf(".%df", len(strings.Split(p.Product.QuoteIncrement, ".")[1])), price)
}

func (p Pattern) Wat(price float64) string {
	return p.precisePrice(price)
}
