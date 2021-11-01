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

	order, err := CreateOrder(pattern.UserID, model.NewMarketEntryOrder(pattern.ProductID, pattern.Size))

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

	entryPrice := util.StringToFloat64(order.ExecutedValue) / util.StringToFloat64(order.Size)
	exitPrice := pattern.GoalPrice(entryPrice)

	sell(wsConn, exitPrice, order.Size, pattern)
}
