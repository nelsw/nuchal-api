package model

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	"github.com/pkg/errors"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
	"time"
)

type Pipe struct {
	wsConn    *ws.Conn
	productID string
}

func NewPipe(productID string) (*Pipe, error) {
	p := &Pipe{productID: productID}
	if err := p.OpenPipe(); err != nil {
		log.Err(err).Stack().Send()
		return nil, err
	}
	return p, nil
}

func (p *Pipe) OpenPipe() error {

	var wsDialer ws.Dialer
	var err error

	if p.wsConn, _, err = wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil); err != nil {
		log.Err(err).Msg("opening ws")
		return err
	}

	log.Trace().Msg("pipe connected")

	if err = p.wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{p.productID}}},
	}); err != nil {
		log.Err(err).Msg("writing ws")
		return err
	}

	log.Trace().Msg("pipe subscribed")

	return nil
}

func (p *Pipe) ClosePipe() error {
	if err := p.wsConn.Close(); err != nil {
		log.Err(err).Msg("closing pipe")
		return err
	}
	return nil
}

// getPrice gets the latest ticker price for the given productId.
func (p *Pipe) getPrice() (float64, error) {

	var receivedMessage cb.Message
	for {
		if err := p.wsConn.ReadJSON(&receivedMessage); err != nil {
			log.Error().
				Err(err).
				Stack().
				Str("productID", p.productID).
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
			Str("productID", p.productID).
			Msg("error getting ticker message from websocket")
		return 0, err
	}

	return util.StringToFloat64(receivedMessage.Price), nil
}

func (p *Pipe) getRate() (Rate, error) {

	end := time.Now().Add(time.Minute)

	var low, high, open, volume float64
	for {

		price, err := p.getPrice()
		if err != nil {
			log.Error().Err(err).Str("productID", p.productID).Msg("price")
			return Rate{}, err
		}

		volume++

		if low == 0 {
			low = price
			high = price
			open = price
		} else if high < price {
			high = price
		} else if low > price {
			low = price
		}

		if time.Now().After(end) {

			rate := NewRate(p.productID, cb.HistoricRate{
				Time:   time.Now().UTC(),
				Low:    low,
				High:   high,
				Open:   open,
				Close:  price,
				Volume: volume,
			})

			rate.log().Info().Msg("rate")

			return rate, nil
		}
	}
}
