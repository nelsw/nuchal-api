package model

import (
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
)

func NewTrade(patternID uint) {
	runTrades(patternID)
}

func NewSell(patternID uint, orderID string) error {
	pattern := FindPatternByID(patternID)
	order, err := GetOrder(pattern, orderID)
	if err != nil {
		return err
	}
	sell(order, pattern)
	return nil
}

func runTrades(patternID uint) {
	go tradeIt(patternID)
	select {}
}

func tradeIt(patternID uint) {

	var err error
	var buys int
	var then, that, this Rate
	for {

		pattern := FindPatternByID(patternID)

		if !pattern.Enable {
			pattern.log().Info().Msg("pattern disabled")
			return
		}

		if this, err = rate(pattern.ProductID); err != nil {
			pattern.log().Error().Err(err).Msg("rate")
			then = Rate{}
			that = Rate{}
			continue
		}

		if pattern.MatchesTweezerBottomPattern(then, that, this) {
			go buyBabyBuy(pattern)
			buys++
			if pattern.Bound == buyBound && pattern.Bind >= buys {
				pattern.log().Info().Int("bind", buys).Msg("bound")
			}
		}

		then = that
		that = this
	}
}

func buyBabyBuy(pattern Pattern) {

	pattern.log().Info().Msg("buyOrder")

	order, err := CreateOrder(pattern, pattern.NewMarketEntryOrder())

	if err != nil {
		pattern.log().Error().Err(err).Msg("buyOrder")
		return
	}

	log.Info().
		Uint("userID", pattern.UserID).
		Uint("patternID", pattern.ID).
		Str("productID", pattern.ProductID).
		Str("orderId", order.ID).
		Msg("created order")

	for {
		if err = sellBabySell(order, pattern); err != nil {
			pattern.log().Error().Err(err).Msg("sellOrder")
			break
		}
	}
}

func sell(order cb.Order, pattern Pattern) {
	go sellIt(order, pattern)
	select {}
}

func sellIt(order cb.Order, pattern Pattern) {
	for {
		if err := sellBabySell(order, pattern); err != nil {
			log.Err(err).Stack().Send()
			break
		}
	}
}

func sellBabySell(order cb.Order, pattern Pattern) error {

	entry := util.StringToFloat64(order.ExecutedValue) / util.StringToFloat64(order.Size)

	goal := pattern.GoalPrice(entry)
	loss := pattern.LossPrice(entry)

	pattern.log().Info().
		Float64("entry", entry).
		Float64("goal", goal).
		Float64("loss", loss).
		Str("size", order.Size).
		Msg("sellOrder")

	var err error

	if order, err = CreateOrder(pattern, pattern.StopLossOrder(order.Size, loss)); err != nil {
		pattern.log().Error().Stack().Err(err).Msg("error placing stop loss order")
	} // just keep going, yolo my brolo

	var wsDialer ws.Dialer
	var wsConn *ws.Conn

	if wsConn, _, err = wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil); err != nil {
		log.Error().Err(err).Stack().Msg("opening ws")
		return err
	}

	defer func(wsConn *ws.Conn) {
		if err = wsConn.Close(); err != nil {
			log.Error().Err(err).Stack().Msg("closing ws")
		}
	}(wsConn)

	if err = wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{pattern.ProductID}}},
	}); err != nil {
		log.Error().Err(err).Stack().Msg("writing ws")
		return err
	}

	for {

		var price float64

		if price, err = getPrice(wsConn, pattern.ProductID); err != nil {
			pattern.log().Error().Err(err).Msg("error getting price during sellOrder")

			// can't get price info so create a stop entry order in case the price reaches our goal
			if _, err = CreateOrder(pattern, pattern.NewStopEntryOrder(order.Size, goal)); err != nil {
				pattern.log().Error().Err(err).Msg("error while creating stop entry order")
			}

			return err // can't proceed without price data, return an error and RUN IT AGAIN!
		}

		if price <= loss {
			pattern.log().Info().Float64("price", price).Float64("exit", loss).Msg("price <= loss")
			return nil
		}

		if price < goal {
			continue
		}

		pattern.log().Info().Float64("price", price).Float64("goal", goal).Msg("price >= goal")

		// place a stop loss at our goal and try to find a higher price
		if order, err = CreateOrder(pattern, pattern.StopLossOrder(order.Size, goal)); err != nil {
			pattern.log().Error().Err(err).Float64("price", price).Float64("goal", goal).Msg("error while creating stop loss order")

			// nvm, sellOrder asap - hopefully the price is still higher than our goal

			if order, err = CreateOrder(pattern, pattern.NewMarketExitOrder()); err != nil {
				pattern.log().Error().Err(err).Float64("price", price).Float64("goal", goal).Msg("error while creating market exit order")

				return err // wow, we can't even place a market exit order, RUN IT AGAIN!
			}

			log.Info().
				Uint("userID", pattern.UserID).
				Str("productID", pattern.ProductID).
				Float64("goal", goal).
				Float64("exit", util.StringToFloat64(order.ExecutedValue)/util.StringToFloat64(order.Size)).
				Msg("price >= goal")

			return nil // we sold at market price so no point in returning an error and triggering another sellOrder
		}

		for { // if we're here, we placed a stop loss order at our goal and now we try to find a better price

			var rate Rate

			if rate, err = getRate(wsConn, pattern.ProductID); err != nil {
				log.Error().
					Err(err).
					Uint("userID", pattern.UserID).
					Str("productID", pattern.ProductID).
					Msg("error while getting rate during price climb")
				continue // since we have a stop loss placed, no harm in continuing onto the next rate
			}

			if rate.Low <= goal {
				log.Info().
					Uint("userID", pattern.UserID).
					Str("productID", pattern.ProductID).
					Float64("rate.Low", rate.Low).
					Float64("exit", goal).
					Msg("rate.Low <= goal, sold")
				return nil
			}

			if rate.Close > goal {
				log.Info().
					Uint("userID", pattern.UserID).
					Str("productID", pattern.ProductID).
					Float64("rate.Close", rate.Close).
					Float64("goal", goal).
					Msg("rate.Close > goal, we found a better price!")

				if err = CancelOrder(pattern, order.ID); err != nil {
					log.Error().
						Err(err).
						Uint("userID", pattern.UserID).
						Str("productID", pattern.ProductID).
						Float64("rate.Close", rate.Close).
						Float64("goal", goal).
						Msg("error while canceling order during price climb")
					return nil // since we have a stop loss placed, no harm in continuing onto the next rate
				}

				goal = rate.Close
				if order, err = CreateOrder(pattern, pattern.StopLossOrder(order.Size, goal)); err != nil {
					log.Error().
						Err(err).
						Uint("userID", pattern.UserID).
						Str("productID", pattern.ProductID).
						Msg("error while creating stop loss order")

					// sellOrder asap - we have no stops in place
					var order cb.Order
					if order, err = CreateOrder(pattern, pattern.NewMarketExitOrder()); err != nil {
						log.Error().
							Err(err).
							Uint("userID", pattern.UserID).
							Str("productID", pattern.ProductID).
							Float64("price", price).
							Float64("goal", goal).
							Msg("error while creating market exit order after failing to create stop loss order during sellOrder")
						return err // this is bad, and super rare ... we need to start over
					}

					log.Info().
						Uint("userID", pattern.UserID).
						Str("productID", pattern.ProductID).
						Float64("goal", goal).
						Float64("exit", util.StringToFloat64(order.ExecutedValue)/util.StringToFloat64(order.Size)).
						Msg("sold")

					return nil // we have a nice stop loss in place, no need to try and sellOrder this fill again
				}
			}
		} // else, onto the next rate and hopefully another higher price, woot!
	}
}
