package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
	"time"
)

type Portfolio struct {
	Time      time.Time  `json:"time"`
	Positions []Position `json:"positions"`
	Cash      string     `json:"cash"`
	Crypto    string     `json:"crypto"`
	Value     string     `json:"value"`
	Hold      float64    `json:"hold"`
	Qty       float64    `json:"qty"`
}

type Position struct {
	ID      string     `json:"id"`
	Balance float64    `json:"balance"`
	Hold    float64    `json:"hold"`
	Last    string     `json:"last"`
	Mean    string     `json:"mean"`
	Gross   string     `json:"gross"`
	Net     string     `json:"net"`
	Profit  string     `json:"profit"`
	Fills   []cb.Fill  `json:"fills,omitempty"`
	Orders  []cb.Order `json:"orders,omitempty"`
}

type Posture struct {
}

func GetPortfolio(userID uint) (Portfolio, error) {

	u := FindUserByID(userID)

	var accounts []cb.Account
	var err error

	if accounts, err = u.Client().GetAccounts(); err != nil {
		log.Err(err).Send()
		return Portfolio{}, err
	}

	var cash, qty, crypto, totalBalance, totalHold float64

	var positions []Position
	for _, account := range accounts {

		hold := util.StringToFloat64(account.Hold)
		totalHold += hold

		balance := util.StringToFloat64(account.Balance)
		totalBalance += balance

		if balance == 0.0 && hold == 0.0 {
			continue
		}

		if account.Currency == "USD" {
			cash = balance
			positions = append(positions, Position{ID: "USD", Balance: balance})
			continue
		}

		qty++

		productID := account.Currency + "-USD"

		var fills []cb.Fill
		if fills, err = GetRemainingBuyFills(userID, productID, balance); err != nil {
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
			hold,
			util.StringToUsd(ticker.Price),
			util.FloatToUsd(avg),
			util.FloatToUsd(gross),
			util.FloatToUsd(net),
			util.FloatToUsd(net - gross),
			fills,
			orders,
		})
	}

	return Portfolio{
		time.Now(),
		positions,
		util.FloatToUsd(cash),
		util.FloatToUsd(crypto),
		util.FloatToUsd(cash + crypto),
		totalHold,
		qty,
	}, nil
}
