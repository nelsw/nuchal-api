package view

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/model"
	"nuchal-api/model/product"
	"nuchal-api/util/money"
	"time"
)

type Quote struct {
	ProductID      string `json:"product_id"`
	BaseCurrency   string `json:"base_currency"`
	QuoteCurrency  string `json:"quote_currency"`
	BaseMinSize    string `json:"base_min_size"`
	BaseMaxSize    string `json:"base_max_size"`
	QuoteIncrement string `json:"quote_increment"`

	Price  string    `json:"price"`
	Volume string    `json:"volume_24h"`
	Low    string    `json:"low_24h"`
	High   string    `json:"high_24h"`
	Time   time.Time `json:"time"`

	Trend []int `json:"trend"`
}

func GetQuotes(userID uint) []Quote {

	var quotes []Quote

	u := model.FindUserByID(userID)

	for _, product := range product.ProductArr {
		ticker, err := u.Client().GetTicker(product.ID)
		if err != nil {
			log.Err(err).Send()
			break
		}
		quotes = append(quotes, NewQuote(product, ticker))
	}

	return quotes
}

func NewQuote(product cb.Product, ticker cb.Ticker) Quote {
	var trend []int
	return Quote{
		ProductID:      product.ID,
		BaseCurrency:   product.BaseCurrency,
		QuoteCurrency:  product.QuoteCurrency,
		BaseMinSize:    product.BaseMinSize,
		BaseMaxSize:    product.BaseMaxSize,
		QuoteIncrement: money.StringToDecimal(product.QuoteIncrement),
		Price:          money.StringToUsd(ticker.Price),
		Volume:         money.StringToDecimal(string(ticker.Volume)),
		Time:           ticker.Time.Time(),
		Trend:          trend,
	}
}
