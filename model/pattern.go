package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math"
	"nuchal-api/db"
	"nuchal-api/util"
)

// Pattern defines the criteria for matching rates and placing orders.
type Pattern struct {

	// UintModel
	UintModel

	// UserId
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

	Product *Product `json:"product"`

	User *User `json:"-"`

	Projection Projection `json:"projection" gorm:"-"`
}

type Projection struct {
	Buy  float64 `json:"buy"`
	Sell float64 `json:"sell"`
	Fees float64 `json:"fees"`
	ROI  float64 `json:"roi"`
}

func init() {
	db.Migrate(&Pattern{})
}

func (p Pattern) Logger() *zerolog.Logger {
	logger := log.
		With().
		Uint("userID", p.UserID).
		Uint("patternID", p.ID).
		Str("productID", p.Currency()).
		Logger()
	return &logger
}

func (p Pattern) Currency() string {
	return p.Product.ID
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

func GetPatterns(userID uint) *[]Pattern {

	var patterns []Pattern

	db.Resolve().
		Preload("Product").
		Preload("User").
		Where("user_id = ?", userID).
		Find(&patterns)

	var newPatterns []Pattern
	for _, pattern := range patterns {
		_ = pattern.Product.initPrice()
		buy := pattern.Product.Price * pattern.Size
		sell := (buy * pattern.Target) + buy
		fees := (buy * pattern.User.Maker) + (sell * pattern.User.Taker)
		pattern.Projection.Buy = util.StringToFloat64(pattern.Product.precise(buy))
		pattern.Projection.Sell = util.StringToFloat64(pattern.Product.precise(sell))
		pattern.Projection.Fees = util.StringToFloat64(pattern.Product.precise(fees))
		pattern.Projection.ROI = util.StringToFloat64(pattern.Product.precise(sell - buy - fees))
		newPatterns = append(newPatterns, pattern)
	}

	return &newPatterns
}

func FindPattern(id uint) Pattern {
	var pattern Pattern
	db.Resolve().
		Preload("Product").
		Where("id = ?", id).
		Find(&pattern)
	return pattern
}

func (p Pattern) NewMarketEntryOrder() cb.Order {
	return cb.Order{
		ProductID: p.Currency(),
		Side:      "buyOrder",
		Size:      util.FloatToDecimal(p.Size),
		Type:      "market",
	}
}

func (p Pattern) NewMarketExitOrder() cb.Order {
	return cb.Order{
		ProductID: p.Currency(),
		Side:      "sellOrder",
		Size:      util.FloatToDecimal(p.Size),
		Type:      "market",
	}
}

func (p Pattern) NewStopEntryOrder(size string, price float64) cb.Order {
	return cb.Order{
		Price:     p.Product.precise(price),
		ProductID: p.Currency(),
		Side:      "sellOrder",
		Size:      size,
		Type:      "limit",
		StopPrice: p.Product.precise(price),
		Stop:      "entry",
	}
}

func (p Pattern) StopLossOrder(size string, price float64) cb.Order {
	return cb.Order{
		Price:     p.Product.precise(price),
		ProductID: p.Currency(),
		Side:      "sellOrder",
		Size:      size,
		Type:      "limit",
		StopPrice: p.Product.precise(price),
		Stop:      "loss",
	}
}
