package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
	"sort"
)

const (
	zero = "0.0000000000000000"
)

type Position struct {
	ProductID          string    `json:"product_id"`
	Balance            float64   `json:"balance"`
	Hold               float64   `json:"hold"`
	CurrentTickerPrice string    `json:"current_ticker_price"`
	AverageFillPrice   string    `json:"average_fill_price"`
	Gross              string    `json:"gross"`
	Net                string    `json:"net"`
	Profit             string    `json:"profit"`
	Fills              []cb.Fill `json:"fills,omitempty"`
}

func GetPositions(userID uint) ([]Position, error) {

	u := FindUserByID(userID)

	accounts, err := u.Client().GetAccounts()
	if err != nil {
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
			})
			continue
		}

		productID := account.Currency + "-USD"

		fills, err := getRecentBuyFills(u, productID)
		if err != nil {
			log.Err(err).Send()
			return nil, err
		}

		var sum float64
		for _, fill := range fills {
			sum += util.StringToFloat64(fill.Price)
		}
		avg := sum / float64(len(fills))

		ticker, err := u.Client().GetTicker(productID)
		if err != nil {
			log.Err(err).Send()
			return nil, err
		}

		balance := util.StringToFloat64(account.Balance)
		price := util.StringToFloat64(ticker.Price)
		gross := price * balance
		net := gross*u.Taker + gross

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
		})
	}

	return positions, nil
}

func getRecentBuyFills(u User, productID string) ([]cb.Fill, error) {

	cursor := u.Client().ListFills(cb.ListFillsParams{ProductID: productID})

	var newFills, allFills []cb.Fill
	for cursor.HasMore {

		if err := cursor.NextPage(&newFills); err != nil {
			return nil, err
		}

		for _, chunk := range newFills {
			allFills = append(allFills, chunk)
		}
	}

	sort.SliceStable(allFills, func(i, j int) bool {
		return allFills[i].CreatedAt.Time().Before(allFills[j].CreatedAt.Time())
	})

	var buys, sells []cb.Fill

	for _, fill := range allFills {
		if fill.Side == "buy" {
			buys = append(buys, fill)
		} else {
			sells = append(sells, fill)
		}
	}

	qty := util.MinInt(len(buys), len(sells))
	result := buys[qty:]

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].CreatedAt.Time().After(result[j].CreatedAt.Time())
	})

	return result, nil
}
