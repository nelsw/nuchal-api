package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/service"
	"nuchal-api/util"
)

const (
	zero = "0.0000000000000000"
)

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

func GetPositions(userID uint) ([]Position, error) {

	u := FindUserByID(userID)

	var accounts []cb.Account
	var err error

	if accounts, err = u.Client().GetAccounts(); err != nil {
		log.Err(err).Send()
		return nil, err
	}

	var positions []Position
	for _, account := range accounts {

		if account.Balance == zero && account.Hold == zero {
			continue
		}

		if account.Currency == "USD" {
			positions = append(positions, Position{
				"USD",
				util.StringToFloat64(account.Balance),
				0,
				"$1.00",
				"$1.00",
				util.StringToUsd(account.Balance),
				util.StringToUsd(account.Balance),
				util.StringToUsd(account.Balance),
				nil,
				nil,
			})
			continue
		}

		productID := account.Currency + "-USD"

		var fills []cb.Fill
		if fills, err = service.GetRemainingBuyFills(userID, productID); err != nil {
			log.Err(err).Send()
			return nil, err
		}

		var orders []cb.Order
		if orders, err = service.GetOrders(userID, productID); err != nil {
			log.Err(err).Send()
			return nil, err
		}

		var ticker cb.Ticker
		if ticker, err = u.Client().GetTicker(productID); err != nil {
			log.Err(err).Send()
			return nil, err
		}

		var sum float64
		for _, fill := range fills {
			sum += util.StringToFloat64(fill.Price)
		}

		balance := util.StringToFloat64(account.Balance)
		gross := util.StringToFloat64(ticker.Price) * balance
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

	return positions, nil
}
