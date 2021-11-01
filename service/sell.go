package service

import (
	ws "github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"nuchal-api/model"
)

func sell(wsConn *ws.Conn, exitPrice float64, size string, pattern model.Pattern) {

	log.Info().
		Uint("userID", pattern.UserID).
		Str("productID", pattern.ProductID).
		Float64("exitPrice", exitPrice).
		Str("size", size).
		Msg("sell")

	for {

		var lastPrice float64
		var err error

		if lastPrice, err = getPrice(wsConn, pattern.ProductID); err != nil {

			log.Error().
				Err(err).
				Uint("userID", pattern.UserID).
				Str("productID", pattern.ProductID).
				Msg("error getting price during sell")

			if _, err := CreateOrder(pattern.UserID, model.NewStopEntryOrder(pattern.ProductID, size, exitPrice)); err != nil {
				log.Error().
					Err(err).
					Uint("userID", pattern.UserID).
					Str("productID", pattern.ProductID).
					Msg("error while creating stop entry order during sell")
			}
			return
		}

		if lastPrice < exitPrice {
			continue
		}

		stopLossOrder := model.StopLossOrder(pattern.ProductID, size, lastPrice)

		if stopLossOrder, err = CreateOrder(pattern.UserID, stopLossOrder); err != nil {
			log.Error().
				Err(err).
				Uint("userID", pattern.UserID).
				Str("productID", pattern.ProductID).
				Msg("error while creating stop loss order")
			return
		}

		for {

			var rate model.Rate

			if rate, err = getRate(wsConn, pattern.ProductID); err != nil {
				log.Error().
					Err(err).
					Uint("userID", pattern.UserID).
					Str("productID", pattern.ProductID).
					Msg("error while getting rate during stop loss climb")
				return
			}

			if rate.Low <= exitPrice {
				return // stop loss executed
			}

			if rate.Close > exitPrice {

				log.Info().Msg("found better price!")

				if err = CancelOrder(pattern.UserID, stopLossOrder.ID); err != nil {
					log.Error().
						Err(err).
						Uint("userID", pattern.UserID).
						Str("productID", pattern.ProductID).
						Msg("error while canceling order")
					return
				}

				exitPrice = rate.Close
				if stopLossOrder, err = CreateOrder(pattern.UserID, model.StopLossOrder(pattern.ProductID, size, exitPrice)); err != nil {
					log.Error().
						Err(err).
						Uint("userID", pattern.UserID).
						Str("productID", pattern.ProductID).
						Msg("error while creating stop loss order")
					return
				}
			}
		}
	}
}
