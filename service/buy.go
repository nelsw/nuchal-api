package service

import (
	ws "github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"nuchal-api/model"
	"nuchal-api/util"
)

func buy(wsConn *ws.Conn, userID uint, pattern model.Pattern) {

	log.Info().
		Uint("userID", userID).
		Str("productId", pattern.ProductID).
		Msg("buy")

	order, err := CreateOrder(userID, model.NewMarketEntryOrder(pattern.ProductID, pattern.Size))

	if err != nil {
		log.Error().
			Err(err).
			Uint("userID", userID).
			Str("productId", pattern.ProductID).
			Msg("error buying")
		return
	}

	log.Info().
		Uint("userID", userID).
		Str("productId", pattern.ProductID).
		Str("orderId", order.ID).
		Msg("created order")

	entryPrice := util.StringToFloat64(order.ExecutedValue) / util.StringToFloat64(order.Size)
	exitPrice := pattern.GoalPrice(entryPrice)

	sell(wsConn, userID, exitPrice, order.Size, pattern)
}
