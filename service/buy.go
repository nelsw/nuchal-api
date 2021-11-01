package service

import (
	ws "github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"nuchal-api/model"
	"nuchal-api/util"
)

func buy(wsConn *ws.Conn, pattern model.Pattern) {

	log.Info().
		Uint("userID", pattern.UserID).
		Str("productId", pattern.ProductID).
		Msg("buy")

	order, err := CreateOrder(pattern.UserID, pattern.NewMarketEntryOrder())

	if err != nil {
		log.Error().
			Err(err).
			Uint("userID", pattern.UserID).
			Str("productId", pattern.ProductID).
			Msg("error buying")
		return
	}

	log.Info().
		Uint("userID", pattern.UserID).
		Str("productId", pattern.ProductID).
		Str("orderId", order.ID).
		Msg("created order")

	entry := util.StringToFloat64(order.ExecutedValue) / util.StringToFloat64(order.Size)

	for {
		if err := sell(wsConn, entry, order.Size, pattern); err == nil {
			break
		}
	}
}
