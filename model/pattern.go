package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math"
	"nuchal-api/db"
	"nuchal-api/util"
)

type BoundType int

const (
	notBound BoundType = iota
	buyBound
	holdBound
)

type StatusType int

const (
	unknown StatusType = iota
	active
	errored
	inactive
)

// Pattern defines the criteria for matching rates and placing orders.
type Pattern struct {

	// UintModel
	UintModel

	// UserID
	UserID uint `json:"user_id"`

	// Currency is concatenation of two currencies. e.g. BTC-USD
	ProductID string `json:"product_id" gorm:"product_id"`

	// Target is a percentage used to produce the goal sellOrder price from the entry buyOrder price.
	Target float64 `json:"target"`

	// Tolerance is a percentage used to derive a limit sellOrder price from the entry buyOrder price.
	Tolerance float64 `json:"tolerance"`

	// Size is the amount of the transaction, using the ProductMap native quote increment.
	Size float64 `json:"size"`

	// Delta is the size of an acceptable difference between tweezer bottom candlesticks.
	Delta float64 `json:"delta"`

	// Bound is the context to which this strategy looks to achieve so that it can break.
	Bound BoundType `json:"bound"`

	// Bind is a numerical value which gets applied to the Bound.
	Bind int `json:"bind"`

	// Enable is a flag that allows the system to bind, get bound, and break.
	Enable bool `json:"enable"`

	Status StatusType `json:"status"`

	Product Product `json:"product"`

	User User `json:"-"`

	Projection Projection `json:"projection" gorm:"-"`
}

func init() {
	db.Migrate(&Pattern{})
}

func (p Pattern) log() *zerolog.Logger {
	logger := log.
		With().
		Str(`productID`, p.ProductID).
		Logger()
	return &logger
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
	if p.ID > 0 {
		db.Resolve().Save(p)
	} else {
		db.Resolve().Create(p)
	}
}

func DeletePattern(patternID uint) {
	db.Resolve().Delete(&Pattern{}, patternID)
}

func FindPatternByID(patternID uint) Pattern {
	var pattern Pattern
	db.Resolve().
		Preload("Product").
		Preload("User").
		Where("id = ?", patternID).
		Find(&pattern)
	return pattern
}

func FindFirstPatternByProductID(user User, productID string) Pattern {
	var pattern Pattern

	db.Resolve().
		Preload("User").
		Preload("Product").
		Where("product_id = ?", productID).
		First(&pattern)

	if &pattern == (&Pattern{}) {

		var product Product
		db.Resolve().
			Where("product_id = ?", productID).
			First(&product)

		pattern = Pattern{
			UserID:    user.ID,
			ProductID: productID,
			Target:    (user.Taker + user.Maker) * 3,
			Tolerance: .1,
			Enable:    true,
			Product:   product,
			User:      user,
		}
	}

	return pattern
}

func GetPatterns(userID uint) []Pattern {

	var patterns []Pattern

	db.Resolve().
		Preload("Product").
		Preload("User").
		Where("user_id = ?", userID).
		Find(&patterns)

	var newPatterns []Pattern
	for _, pattern := range patterns {

		var buy float64
		if buy = pattern.Product.Posture.Price * pattern.Size; math.IsNaN(buy) {
			buy = 0
		}

		var sell float64
		if sell = (buy * pattern.Target) + buy; math.IsNaN(sell) {
			sell = 0
		}

		var fees float64
		if fees = (buy * pattern.User.Maker) + (sell * pattern.User.Taker); math.IsNaN(fees) {
			fees = 0
		}
		projection := Projection{
			Buy:  buy,
			Sell: sell,
			Fees: fees,
		}

		projection.setValues(pattern.Product.precise)

		pattern.Projection = projection

		newPatterns = append(newPatterns, pattern)
	}

	return newPatterns
}

func FindPattern(id uint) Pattern {
	var pattern Pattern
	db.Resolve().
		Preload("Product").
		Preload("User").
		Where("id = ?", id).
		Find(&pattern)
	return pattern
}

func (p Pattern) NewMarketEntryOrder() cb.Order {
	return p.Product.NewMarketEntryOrder(util.FloatToDecimal(p.Size))
}

func (p Pattern) NewMarketExitOrder() cb.Order {
	return p.Product.NewMarketExitOrder(util.FloatToDecimal(p.Size))
}

func (p Pattern) NewStopEntryOrder(size string, price float64) cb.Order {
	return p.Product.NewStopEntryOrder(size, price)
}

func (p Pattern) StopLossOrder(size string, price float64) cb.Order {
	return p.Product.StopLossOrder(size, price)
}
