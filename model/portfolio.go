package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
	"time"
)

const (
	zero = "0.0000000000000000"
)

type Portfolio struct {
	Time      time.Time  `json:"time"`
	Positions []Position `json:"positions"`
	Cash      string     `json:"cash"`
	Crypto    string     `json:"crypto"`
}

type Position struct {
	ProductID          string     `json:"product_id"`
	Balance            float64    `json:"balance"`
	Hold               float64    `json:"hold"`
	CurrentTickerPrice string     `json:"current_ticker_price"`
	AverageFillPrice   string     `json:"average_fill_price"`
	Gross              string     `json:"gross"`
	Net                string     `json:"net"`
	Profit             string     `json:"profit"`
	Fills              []cb.Fill  `json:"fills,omitempty"`
	Orders             []cb.Order `json:"orders"`
}

func GetPortfolio(userID uint) (Portfolio, error) {

	u := FindUserByID(userID)

	var accounts []cb.Account
	var err error

	if accounts, err = u.Client().GetAccounts(); err != nil {
		log.Err(err).Send()
		return Portfolio{}, err
	}

	var cash string
	var crypto float64

	var positions []Position
	for _, account := range accounts {

		if account.Balance == zero && account.Hold == zero {
			continue
		}

		balance := util.StringToFloat64(account.Balance)

		if account.Currency == "USD" {
			cash = util.StringToUsd(account.Balance)
			positions = append(positions, Position{ProductID: "USD", Balance: balance})
			continue
		}

		productID := account.Currency + "-USD"

		var fills []cb.Fill
		if fills, err = GetRemainingBuyFills(userID, productID); err != nil {
			log.Err(err).Send()
			return Portfolio{}, err
		}

		var orders []cb.Order
		if orders, err = GetOrders(userID, productID); err != nil {
			log.Err(err).Send()
			return Portfolio{}, err
		}

		var ticker cb.Ticker
		if ticker, err = u.Client().GetTicker(productID); err != nil {
			log.Err(err).Send()
			return Portfolio{}, err
		}

		var sum float64
		for _, fill := range fills {
			sum += util.StringToFloat64(fill.Price)
		}

		gross := util.StringToFloat64(ticker.Price) * balance
		crypto += gross
		net := gross*u.Taker + gross
		avg := sum / float64(len(fills))

		positions = append(positions, Position{
			productID,
			balance,
			util.StringToFloat64(account.Hold),
			util.StringToUsd(ticker.Price),
			util.FloatToUsd(avg),
			util.FloatToUsd(gross),
			util.FloatToUsd(net),
			util.FloatToUsd(net - (avg * balance)),
			fills,
			orders,
		})
	}

	return Portfolio{time.Now(), positions, cash, util.FloatToUsd(crypto)}, nil
}
