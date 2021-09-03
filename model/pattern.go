package model

import (
	"gorm.io/gorm"
	"math"
	"nuchal-api/db"
	"nuchal-api/util"
)

// Pattern defines the criteria for matching rates and placing orders.
type Pattern struct {
	gorm.Model

	// UserId
	UserID uint `json:"user_id"`

	// ProductID is concatenation of two currencies. e.g. BTC-USD
	ProductID string `json:"product_id"`

	// Gain is a percentage used to produce the goal sell price from the entry buy price.
	Gain float64 `json:"gain"`

	// Loss is a percentage used to derive a limit sell price from the entry buy price.
	Loss float64 `json:"loss"`

	// Size is the amount of the transaction, using the ProductMap native quote increment.
	Size float64 `json:"size"`

	// Delta is the size of an acceptable difference between tweezer bottom candlesticks.
	Delta float64 `json:"delta"`
}

func init() {
	db.Migrate(&Pattern{})
}

func (p *Pattern) GoalPrice(price float64) float64 {
	return price + (price * p.Gain)
}

func (p *Pattern) LossPrice(price float64) float64 {
	return price - (price * p.Loss)
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
		db.Resolve().Save(&p)
	} else {
		db.Resolve().Create(&p)
	}
}

func GetPattern(userID uint, productID string) Pattern {

	pattern := Pattern{}

	db.Resolve().
		Where("user_id = ?", userID).
		Where("product_id = ?", productID).
		Find(&pattern)

	if pattern.ProductID != productID {

		user := FindUserByID(userID)
		product := ProductMap[productID]

		fees := user.Maker + user.Taker
		gain := fees * 2
		loss := fees * 10
		size := util.StringToFloat64(product.BaseMinSize)
		delta := util.StringToFloat64(product.QuoteIncrement)

		p := &Pattern{
			ProductID: productID,
			Gain:      gain,
			Loss:      loss,
			Size:      size,
			Delta:     delta,
			UserID:    userID,
		}

		db.Resolve().Create(p)

		return *p
	}

	return pattern
}
