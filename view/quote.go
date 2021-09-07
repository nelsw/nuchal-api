package view

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/model"
	"nuchal-api/util"
	"time"
)

type Quote struct {
	ProductID      string `json:"product_id"`
	BaseCurrency   string `json:"base_currency"`
	QuoteCurrency  string `json:"quote_currency"`
	BaseMinSize    string `json:"base_min_size"`
	BaseMaxSize    string `json:"base_max_size"`
	QuoteIncrement string `json:"quote_increment"`

	Price     string    `json:"price"`
	Volume24h string    `json:"volume_24h"`
	Volume30d string    `json:"volume_30d"`
	Open24h   string    `json:"open_24h"`
	Low       string    `json:"low_24h"`
	High      string    `json:"high_24h"`
	Time      time.Time `json:"time"`

	Trend []int `json:"trend"`

	Direction     string `json:"direction"`
	ChangePercent string `json:"change_percent"`
	ChangePrice   string `json:"change_price"`
}

func GetQuotes(userID uint) []Quote {

	var quotes []Quote

	u := model.FindUserByID(userID)

	for _, product := range model.ProductArr {
		ticker, err := u.Client().GetTicker(product.ID())
		if err != nil {
			log.Err(err).Send()
			break
		}
		quotes = append(quotes, NewQuote(product.ID(), product.Product, ticker))
	}

	return quotes
}

func NewQuote(ID string, product cb.Product, ticker cb.Ticker) Quote {
	var trend []int
	return Quote{
		ProductID:      ID,
		BaseCurrency:   product.BaseCurrency,
		QuoteCurrency:  product.QuoteCurrency,
		BaseMinSize:    product.BaseMinSize,
		BaseMaxSize:    product.BaseMaxSize,
		QuoteIncrement: util.StringToDecimal(product.QuoteIncrement),
		Price:          util.StringToUsd(ticker.Price),
		Volume24h:      util.StringToDecimal(string(ticker.Volume)),
		Open24h:        "",
		Volume30d:      "",
		Time:           ticker.Time.Time(),
		Trend:          trend,
		Direction:      "",
		ChangePrice:    "",
		ChangePercent:  "",
	}
}
