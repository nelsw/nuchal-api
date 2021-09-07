package model

import (
	"gorm.io/gorm"
	"math"
	"nuchal-api/db"
)

// Pattern defines the criteria for matching rates and placing orders.
type Pattern struct {
	gorm.Model

	// UserId
	UserID uint `json:"user_id"`

	// ProductID is concatenation of two currencies. e.g. BTC-USD
	ProductID string `json:"product_id"`

	// Target is a percentage used to produce the goal sell price from the entry buy price.
	Target float64 `json:"target"`

	// Tolerance is a percentage used to derive a limit sell price from the entry buy price.
	Tolerance float64 `json:"tolerance"`

	// Size is the amount of the transaction, using the ProductMap native quote increment.
	Size float64 `json:"size"`

	// Delta is the size of an acceptable difference between tweezer bottom candlesticks.
	Delta float64 `json:"delta"`

	// Bind is represents the time of day when this strategy should activate.
	// Value of bind is the total amount of milliseconds totaling hour and minutes.
	Bind int64 `json:"bind"`

	// Bound is the context to which this strategy looks to achieve so that it can break.
	// Values include buys, sells, holds, hours, minutes.
	Bound string `json:"bound"`

	// Break is a numerical value which gets applied to the Bound.
	Break int `json:"break"`

	// Enable is a flag that allows the system to bind, get bound, and break.
	Enable bool `json:"enable"`
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

func GetPatterns(userID uint) []Pattern {
	var patterns []Pattern
	db.Resolve().
		Where("user_id = ?", userID).
		Find(&patterns)
	return patterns
}

func GetPattern(userID uint, productID string) Pattern {

	pattern := Pattern{}

	db.Resolve().
		Where("user_id = ?", userID).
		Where("product_id = ?", productID).
		Find(&pattern)

	return pattern
}
