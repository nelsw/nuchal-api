package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"nuchal-api/db"
	"nuchal-api/util"
)

type SessionOutcome int

const (
	unknownOutcome SessionOutcome = iota
	errorOutcome
	goalOutcome
	gainOutcome
	lossOutcome
	buyOutcome
	disabledOutcome
	boundOutcome
)

type Sessions struct {
	Buys  []BuySession  `json:"buys"`
	Sells []SellSession `json:"sells"`
}

type Session struct {
	gorm.Model
	UserID    uint            `json:"user_id"`
	ProductID string          `json:"product_id"`
	Size      float64         `json:"size"`
	Results   []SessionResult `json:"results" gorm:"polymorphic:Session;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type BuySession struct {
	Session
	PatternID uint `json:"pattern_id"`
}

type SellSession struct {
	Session
	Price float64 `json:"price"`
	Goal  float64 `json:"goal"`
	Even  float64 `json:"even"`
	Loss  float64 `json:"loss"`
	Maker float64 `json:"maker"`
	Taker float64 `json:"taker"`
}

type SessionResult struct {
	gorm.Model
	SessionID   uint           `json:"session_id"`
	SessionType string         `json:"session_type"`
	Error       string         `json:"error"`
	Price       float64        `json:"price"`
	Outcome     SessionOutcome `json:"outcome" gorm:"bigint"`
}

func init() {
	db.Migrate(&BuySession{})
	db.Migrate(&SellSession{})
	db.Migrate(&SessionResult{})
}

func GetSessions(userID uint) Sessions {
	var sessions Sessions
	db.Resolve().Preload("Results").Where("user_id = ?", userID).Find(&sessions.Buys)
	db.Resolve().Preload("Results").Where("user_id = ?", userID).Find(&sessions.Sells)
	return sessions
}

/*
	buy session methods
*/

func StartBuySession(patternID uint) {

	session := &BuySession{PatternID: patternID}

	db.Resolve().Create(&session)

	go session.buy()
}

func (s *BuySession) log() *zerolog.Logger {
	logger := log.
		With().
		Str("productID", s.ProductID).
		Float64("size", s.Size).
		Logger()
	return &logger
}

func (s *BuySession) buy() {

	var pipe *Pipe
	var err error

	if pipe, err = NewPipe(s.ProductID); err != nil {
		s.errorResult(s.log(), err)
		return
	}

	defer func(pipe *Pipe) {
		if err := pipe.ClosePipe(); err != nil {
			s.errorResult(s.log(), err)
		}
	}(pipe)

	var then, that, this Rate
	for {

		pattern := FindPatternByID(s.PatternID)

		if !pattern.Enable {

			s.Results = append(s.Results, SessionResult{SessionID: s.ID, Outcome: disabledOutcome})
			db.Resolve().Save(s)
			return
		}

		if pattern.Bound == buyBound && s.getBuyCount() >= pattern.Bind {
			s.Results = append(s.Results, SessionResult{SessionID: s.ID, Outcome: boundOutcome})
			db.Resolve().Save(s)
			return
		}

		if this, err = pipe.getRate(); err != nil {
			s.errorResult(s.log(), err)
			return
		}

		if !pattern.MatchesTweezerBottomPattern(then, that, this) {
			continue
		}

		var price, size float64
		if price, size, err = pattern.placeMarketEntryOrder(); err != nil {
			s.errorResult(s.log(), err)
			return
		}

		go startSellSession(price, size, pattern)

		then = that
		that = this
	}

}

func (s *BuySession) getBuyCount() int64 {
	var count int64
	db.Resolve().
		Model(&SessionResult{}).
		Where("session_id = ?", s.ID).
		Where("session_outcome = ?", buyOutcome).
		Count(&count)
	return count
}

/*
	sell session methods
*/
func StartSellSession(price, size float64, productID string) {
	startSellSession(price, size, FindFirstPatternByProductID(productID))
}

func startSellSession(price, size float64, pattern Pattern) {

	session := &SellSession{
		Session: Session{
			UserID:    pattern.UserID,
			ProductID: pattern.ProductID,
			Size:      size,
		},
		Price: price,
		Goal:  pattern.GoalPrice(price),
		Even:  pattern.EvenPrice(price),
		Loss:  pattern.LossPrice(price),
		Maker: pattern.User.Maker,
		Taker: pattern.User.Taker,
	}

	db.Resolve().Create(&session)

	session.sell()
}

func (s *SellSession) Run(event *zerolog.Event, level zerolog.Level, msg string) {
	event.
		Str("productID", s.ProductID).
		Float64("price", s.Price).
		Float64("size", s.Size).
		Float64("goal", s.Goal).
		Float64("even", s.Even).
		Float64("loss", s.Loss)
}

func (s *SellSession) log() *zerolog.Logger {
	logger := log.Hook(s)
	return &logger
}

func (s *SellSession) sell() {

	s.log().Trace().Msg("sell")

	var orderID string
	var pipe *Pipe
	var err error

	if orderID, err = s.anchor(); err != nil {
		s.errorResult(s.log(), err)
		return
	}

	if pipe, err = NewPipe(s.ProductID); err != nil {
		s.errorResult(s.log(), err)
		return
	}

	defer func(pipe *Pipe) {
		if err := pipe.ClosePipe(); err != nil {
			s.errorResult(s.log(), err)
		}
	}(pipe)

	for {

		var price float64

		if price, err = pipe.getPrice(); err != nil {
			s.log().Trace().Msg("error getting price to find goal")
			s.errorResult(s.log(), err)
			return
		}

		if price <= s.Loss {
			s.log().Trace().Msg("price <= goal")
			s.lossResult()
			return
		}

		if price < s.Goal {
			s.log().Trace().Msg("price < goal")
			continue
		}

		if err = s.cancelOrder(orderID); err != nil {
			s.log().Trace().Msg("error canceling stop loss to anchor for goal")
			s.errorResult(s.log(), err)
			return
		}

		if orderID, err = s.anchorOrExit(); err != nil {
			s.log().Trace().Msg("error anchoring for goal")
			s.errorResult(s.log(), err)
			return
		}

		/*
			TIME TO CLIMB BABY WOOOOOOT
		*/

		for {

			var rate Rate
			l := s.log().Hook(rate)
			l.Trace().Send()

			if rate, err = pipe.getRate(); err != nil {
				l.Trace().Msg("error getting price to find gain")
				s.errorResult(&l, err)
				return
			}

			if rate.Low <= price {
				l.Trace().Msg("rate.low <= price")
				s.goalResult()
				return
			}

			if rate.Close > price {

				if err = s.cancelOrder(orderID); err != nil {
					l.Trace().Msg("error canceling stop loss to anchor for gain")
					s.errorResult(&l, err)
					return
				}

				if orderID, err = s.anchorOrExit(); err != nil {
					l.Trace().Msg("error anchoring for gain")
					s.errorResult(&l, err)
					return
				}

				// the new price to beat
				price = rate.Close
			}
		}
	}
}

func (s *SellSession) anchor() (string, error) {
	u := FindUserByID(s.UserID)
	order, err := u.Client().CreateOrder(&cb.Order{
		ProductID: s.ProductID,
		Price:     util.FloatToDecimal(s.Price),
		Side:      "sell",
		Size:      util.FloatToDecimal(s.Size),
		Type:      "limit",
		StopPrice: util.FloatToDecimal(s.Price),
		Stop:      "loss",
	})
	if err != nil {
		return "", err
	}
	return order.ID, nil
}

func (s *SellSession) anchorOrExit() (string, error) {

	orderID, err := s.anchor()
	if err == nil {
		return orderID, nil
	}

	u := FindUserByID(s.UserID)
	_, err = u.Client().CreateOrder(&cb.Order{
		ProductID: s.ProductID,
		Side:      "sell",
		Size:      util.FloatToDecimal(s.Size),
		Type:      "market",
	})
	return "", err
}

func (s *SellSession) cancelOrder(orderID string) error {
	u := FindUserByID(s.UserID)
	return u.Client().CancelOrder(orderID)
}
func (s *SellSession) lossResult() {
	s.log().Info().Msg("loss")
	s.Results = append(s.Results, SessionResult{SessionID: s.ID, Price: s.Loss, Outcome: lossOutcome})
	db.Resolve().Save(s)
}

func (s *SellSession) goalResult() {
	s.log().Info().Msg("goal")
	s.Results = append(s.Results, SessionResult{SessionID: s.ID, Price: s.Goal, Outcome: goalOutcome})
	db.Resolve().Save(s)
}

func (s *SellSession) gainResult(price float64) {
	s.log().Info().Msg("gain")
	s.Results = append(s.Results, SessionResult{SessionID: s.ID, Price: price, Outcome: gainOutcome})
	db.Resolve().Save(s)
}

func (s *SellSession) errorResult(logger *zerolog.Logger, err error) {
	logger.Err(err).Send()
	s.Results = append(s.Results, SessionResult{
		SessionID: s.ID,
		Error:     err.Error(),
		Outcome:   errorOutcome,
	})
	db.Resolve().Save(s)
}

func (s *BuySession) errorResult(logger *zerolog.Logger, err error) {
	logger.Err(err).Send()
	s.Results = append(s.Results, SessionResult{
		SessionID: s.ID,
		Error:     err.Error(),
		Outcome:   errorOutcome,
	})
	db.Resolve().Save(s)
}
