package service

import (
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/model"
)

// todo - implement with pattern conditions like bind, break, etc
func trade(pattern model.Pattern) error {

	log.Info().
		Uint("userID", pattern.UserID).
		Str("productId", pattern.ProductID).
		Msg("creating trades")

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Error().
			Err(err).
			Uint("userID", pattern.UserID).
			Str("productID", pattern.ProductID).
			Msg("error while opening websocket connection")
		return err
	}

	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			log.Error().
				Err(err).
				Uint("userID", pattern.UserID).
				Str("productID", pattern.ProductID).
				Msg("error closing websocket connection")
		}
	}(wsConn)

	if err := wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{pattern.ProductID}}},
	}); err != nil {
		log.Error().
			Err(err).
			Uint("userID", pattern.UserID).
			Str("productID", pattern.ProductID).
			Msg("error writing message to websocket")
		return err
	}

	var then, that model.Rate
	for {

		this, err := getRate(wsConn, pattern.ProductID)
		if err != nil {

			log.Error().
				Err(err).
				Uint("userID", pattern.UserID).
				Str("productID", pattern.ProductID).
				Msg("error getting rate")

			then = model.Rate{}
			that = model.Rate{}

			continue
		}

		if pattern.MatchesTweezerBottomPattern(then, that, this) {
			go buy(wsConn, pattern)
		}

		then = that
		that = this
	}
}
