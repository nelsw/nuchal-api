package model

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	"github.com/pkg/errors"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
	"time"
)

type Pipe struct {
	wsConn    *ws.Conn
	productID string
}

func (p *Pipe) log() *zerolog.Logger {
	logger := log.Hook(p)
	return &logger
}

func (p *Pipe) Run(event *zerolog.Event, level zerolog.Level, msg string) {
	event.Str("productID", p.productID)
}

func NewPipe(productID string) (*Pipe, error) {
	p := &Pipe{productID: productID}
	if err := p.Open(); err != nil {
		log.Err(err).Stack().Send()
		return nil, err
	}
	return p, nil
}

func (p *Pipe) Open() error {

	var wsDialer ws.Dialer
	var err error

	if p.wsConn, _, err = wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil); err != nil {
		log.Err(err).Msg("opening ws")
		return err
	}

	p.log().Debug().Msg("connected")

	if err = p.wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{p.productID}}},
	}); err != nil {
		log.Err(err).Msg("writing ws")
		return err
	}

	p.log().Debug().Msg("subscribed")

	return nil
}

func (p *Pipe) Close() error {
	if err := p.wsConn.Close(); err != nil {
		p.log().Err(err).Stack().Send()
		return err
	}
	return nil
}

func (p *Pipe) Reopen() error {
	p.log().Debug().Msg("reopening")
	if err := p.Close(); err != nil {
		p.log().Err(err).Stack().Send()
		return err
	}
	if err := p.Open(); err != nil {
		p.log().Err(err).Stack().Send()
		return err
	}
	p.log().Debug().Msg("reopened")
	return nil
}

// getPrice gets the latest ticker price for the given productId.
func (p *Pipe) getPrice() (float64, error) {

	var receivedMessage cb.Message
	for {
		if err := p.wsConn.ReadJSON(&receivedMessage); err != nil {
			p.log().Err(err).Stack().Send()
			return 0, err
		}
		if receivedMessage.Type != "subscriptions" {
			break
		}
	}

	if receivedMessage.Type != "ticker" {
		err := errors.New(fmt.Sprintf("message type != ticker, %v", receivedMessage))
		p.log().Err(err).Stack().Send()
		return 0, err
	}

	return util.StringToFloat64(receivedMessage.Price), nil
}

func (p *Pipe) getRate() (Rate, error) {

	end := time.Now().Add(time.Minute)

	var err error
	var price, low, high, open, volume float64
	for {

		if price, err = p.getPrice(); err != nil {
			p.log().Err(err).Stack().Send()
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

			return rate, nil
		}
	}
}
