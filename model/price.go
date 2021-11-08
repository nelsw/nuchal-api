package model

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	"github.com/pkg/errors"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
)

func GetPrice(pid string) (float64, error) {

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Error().
			Err(err).
			Str("pid", pid).
			Msg("error while opening websocket connection")
		return 0, err
	}

	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			log.Error().
				Err(err).
				Str("pid", pid).
				Msg("error closing websocket connection")
		}
	}(wsConn)

	if err := wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{pid}}},
	}); err != nil {
		log.Error().
			Err(err).
			Str("pid", pid).
			Msg("error writing message to websocket")
		return 0, err
	}

	return getPrice(wsConn, pid)
}

// getPrice gets the latest ticker price for the given productId.
func getPrice(wsConn *ws.Conn, pid string) (float64, error) {

	var receivedMessage cb.Message
	for {
		if err := wsConn.ReadJSON(&receivedMessage); err != nil {
			log.Error().
				Err(err).
				Str("pid", pid).
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
			Str("pid", pid).
			Msg("error getting ticker message from websocket")
		return 0, err
	}

	return util.StringToFloat64(receivedMessage.Price), nil
}
