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
	ID         string      `json:"id"`
	Last       string      `json:"last"`
	Balance    float64     `json:"balance"`
	Hold       float64     `json:"hold"`
	Projection Projection  `json:"projection"`
	Fills      []BuyFill   `json:"fills,omitempty"`
	Orders     []SellOrder `json:"orders,omitempty"`
}

func GetPortfolio(userID uint) (Portfolio, error) {

	u := FindUserByID(userID)

	var accounts []cb.Account
	var err error

	if accounts, err = u.Client().GetAccounts(); err != nil {
		log.Err(err).Stack().Send()
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

		var product Product
		if product, err = FindProductByID(account.Currency + "-USD"); err != nil {
			log.Err(err).Stack().Send()
			return Portfolio{}, err
		}

		var fills []BuyFill
		if fills, err = GetRemainingBuyFills(userID, product.ID, balance); err != nil {
			log.Err(err).Stack().Send()
			return Portfolio{}, err
		}

		var orders []SellOrder
		if orders, err = GetOrders(userID, product); err != nil {
			log.Err(err).Stack().Send()
			return Portfolio{}, err
		}

		var sum, fee float64
		for _, fill := range fills {
			sum += fill.Price
			fee += fill.Fee
		}

		exit := (product.Price * balance) - (product.Price * balance * u.Taker)
		crypto += exit

		projection := Projection{
			Buy:  sum,
			Sell: product.Price * float64(len(fills)),
			Fees: fee,
		}

		projection.setValues(product.precise)

		position := Position{
			ID:         product.ID,
			Last:       "$" + product.precise(product.Price),
			Balance:    balance,
			Hold:       hold,
			Projection: projection,
			Fills:      fills,
			Orders:     orders,
		}

		positions = append(positions, position)
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
