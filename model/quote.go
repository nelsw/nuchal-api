package model

import (
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
)

type Quote struct {
	Product
	Price float64 `json:"price"`
}

func GetQuotes() ([]Quote, error) {

	var pids []string
	productMap := map[string]Product{}
	for _, product := range FindAllProducts() {
		pids = append(pids, product.ID)
		productMap[product.ID] = product
	}

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Error().
			Err(err).
			Stack().
			Strs("pid", pids).
			Msg("error while opening websocket connection")
		return nil, err
	}

	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			log.Error().
				Err(err).
				Stack().
				Strs("pid", pids).
				Msg("error closing websocket connection")
		}
	}(wsConn)

	if err := wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", pids}},
	}); err != nil {
		log.Error().
			Err(err).
			Stack().
			Strs("pid", pids).
			Msg("error writing message to websocket")
		return nil, err
	}

	var quotes []Quote
	for _, pid := range pids {

		price, err := GetPrice(pid)
		if err != nil {
			log.Err(err).
				Stack().
				Str("pid", pid).
				Msg("error getting price")
			continue
		}

		quote := Quote{productMap[pid], price}
		log.Trace().Interface("quote", quote).Send()
		quotes = append(quotes, quote)
	}

	return quotes, nil
}
