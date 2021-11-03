package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
)

type Quote struct {
	Product
	cb.Ticker
}

func GetQuotes(userID uint) []Quote {

	_ = InitProducts(userID)
	var quotes []Quote

	u := FindUserByID(userID)

	for _, product := range ProductArr {
		ticker, err := u.Client().GetTicker(product.BaseCurrency + "-USD")
		if err != nil {
			log.Err(err).Send()
			break
		}
		quotes = append(quotes, Quote{product, ticker})
	}

	return quotes
}
