package model

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	"github.com/pkg/errors"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
)

// getPrice gets the latest ticker price for the given productId.
func getPrice(wsConn *ws.Conn, productID string) (float64, error) {

	var receivedMessage cb.Message
	for {
		if err := wsConn.ReadJSON(&receivedMessage); err != nil {
			log.Error().
				Err(err).
				Stack().
				Str("productID", productID).
				Msg("error reading from websocket")
			return 0, err
		}
		if receivedMessage.Type != "subscriptions" {
			break
		}
	}

	if receivedMessage.Type != "ticker" {
		err := errors.New(fmt.Sprintf("message type != ticker, %v", receivedMessage))
		log.Error().
			Err(err).
			Stack().
			Str("productID", productID).
			Msg("error getting ticker message from websocket")
		return 0, err
	}

	return util.StringToFloat64(receivedMessage.Price), nil
}
