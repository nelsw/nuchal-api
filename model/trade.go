package model

import (
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
)

func NewTrade(patternID uint) {

	//pattern.Logger().Trace().Msg("trade")
	//
	//var wsDialer ws.Dialer
	//wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	//if err != nil {
	//	pattern.Logger().Error().Err(err).Msg("opening ws")
	//	return err
	//}
	//
	//defer func(wsConn *ws.Conn) {
	//	if err = wsConn.Close(); err != nil {
	//		pattern.Logger().Error().Err(err).Msg("closing ws")
	//	}
	//}(wsConn)
	//
	//if err = wsConn.WriteJSON(&cb.Message{
	//	Type:     "subscribe",
	//	Channels: []cb.MessageChannel{{"ticker", []string{pattern.Currency()}}},
	//}); err != nil {
	//	pattern.Logger().Error().Err(err).Msg("writing ws")
	//	return err
	//}

	runTrades(patternID)
}

func runTrades(patternID uint) {
	go newTrade(patternID)
	select {}
}

func newTrade(patternID uint) {

	var err error
	var buys int
	var then, that, this Rate
	for {

		pattern := FindPatternByID(patternID)

		if !pattern.Enable {
			pattern.Logger().Info().Msg("pattern disabled")
			return
		}

		if this, err = rate(pattern.Currency()); err != nil {
			pattern.Logger().Error().Err(err).Msg("rate")
			then = Rate{}
			that = Rate{}
			continue
		}

		if pattern.MatchesTweezerBottomPattern(then, that, this) {
			go buyBabyBuy(pattern)
			buys++
			if pattern.Break >= buys {
				pattern.Logger().Info().Int("bind", buys).Msg("bound")
			}
		}

		then = that
		that = this
	}
}

func buyBabyBuy(pattern Pattern) {

	pattern.Logger().Info().Msg("buy")

	order, err := CreateOrder(pattern, pattern.NewMarketEntryOrder())

	if err != nil {
		pattern.Logger().Error().Err(err).Msg("buy")
		return
	}

	log.Info().
		Uint("userID", pattern.UserID).
		Uint("patternID", pattern.Model.ID).
		Str("productID", pattern.Currency()).
		Str("orderId", order.ID).
		Msg("created order")

	entry := util.StringToFloat64(order.ExecutedValue) / util.StringToFloat64(order.Size)

	for {
		if err = sellBabySell(entry, order.Size, pattern); err == nil {
			pattern.Logger().Error().Err(err).Msg("sell")
			break
		}
	}
}

func sellBabySell(entry float64, size string, pattern Pattern) error {

	goal := pattern.GoalPrice(entry)
	loss := pattern.LossPrice(entry)

	pattern.Logger().Info().
		Float64("entry", entry).
		Float64("goal", goal).
		Float64("loss", loss).
		Str("size", size).
		Msg("sell")

	var order cb.Order
	var err error

	if order, err = CreateOrder(pattern, pattern.StopLossOrder(size, loss)); err != nil {
		pattern.Logger().Error().Err(err).Msg("error placing stop loss order")
	} // just keep going, yolo my brolo

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Error().Err(err).Msg("opening ws")
		return err
	}

	defer func(wsConn *ws.Conn) {
		if err = wsConn.Close(); err != nil {
			log.Error().Err(err).Msg("closing ws")
		}
	}(wsConn)

	if err = wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{pattern.Currency()}}},
	}); err != nil {
		log.Error().Err(err).Msg("writing ws")
		return err
	}

	for {

		var price float64

		if price, err = getPrice(wsConn, pattern.Currency()); err != nil {
			pattern.Logger().Error().Err(err).Msg("error getting price during sell")

			// can't get price info so create a stop entry order in case the price reaches our goal
			if _, err = CreateOrder(pattern, pattern.NewStopEntryOrder(size, goal)); err != nil {
				pattern.Logger().Error().Err(err).Msg("error while creating stop entry order")
			}

			return err // can't proceed without price data, return an error and RUN IT AGAIN!
		}

		if price <= loss {
			pattern.Logger().Info().Float64("price", price).Float64("exit", loss).Msg("price <= loss")
			return nil
		}

		if price < goal {
			continue
		}

		pattern.Logger().Info().Float64("price", price).Float64("goal", goal).Msg("price >= goal")

		// place a stop loss at our goal and try to find a higher price
		if order, err = CreateOrder(pattern, pattern.StopLossOrder(size, goal)); err != nil {
			pattern.Logger().Error().Err(err).Float64("price", price).Float64("goal", goal).Msg("error while creating stop loss order")

			// nvm, sell asap - hopefully the price is still higher than our goal

			if order, err = CreateOrder(pattern, pattern.NewMarketExitOrder()); err != nil {
				pattern.Logger().Error().Err(err).Float64("price", price).Float64("goal", goal).Msg("error while creating market exit order")

				return err // wow, we can't even place a market exit order, RUN IT AGAIN!
			}

			log.Info().
				Uint("userID", pattern.UserID).
				Str("productID", pattern.Currency()).
				Float64("goal", goal).
				Float64("exit", util.StringToFloat64(order.ExecutedValue)/util.StringToFloat64(order.Size)).
				Msg("price >= goal")

			return nil // we sold at market price so no point in returning an error and triggering another sell
		}

		for { // if we're here, we placed a stop loss order at our goal and now we try to find a better price

			var rate Rate

			if rate, err = getRate(wsConn, pattern.Currency()); err != nil {
				log.Error().
					Err(err).
					Uint("userID", pattern.UserID).
					Str("productID", pattern.Currency()).
					Msg("error while getting rate during price climb")
				continue // since we have a stop loss placed, no harm in continuing onto the next rate
			}

			if rate.Low <= goal {
				log.Info().
					Uint("userID", pattern.UserID).
					Str("productID", pattern.Currency()).
					Float64("rate.Low", rate.Low).
					Float64("exit", goal).
					Msg("rate.Low <= goal, sold")
				return nil
			}

			if rate.Close > goal {
				log.Info().
					Uint("userID", pattern.UserID).
					Str("productID", pattern.Currency()).
					Float64("rate.Close", rate.Close).
					Float64("goal", goal).
					Msg("rate.Close > goal, we found a better price!")

				if err = CancelOrder(pattern, order.ID); err != nil {
					log.Error().
						Err(err).
						Uint("userID", pattern.UserID).
						Str("productID", pattern.Currency()).
						Float64("rate.Close", rate.Close).
						Float64("goal", goal).
						Msg("error while canceling order during price climb")
					return nil // since we have a stop loss placed, no harm in continuing onto the next rate
				}

				goal = rate.Close
				if order, err = CreateOrder(pattern, pattern.StopLossOrder(size, goal)); err != nil {
					log.Error().
						Err(err).
						Uint("userID", pattern.UserID).
						Str("productID", pattern.Currency()).
						Msg("error while creating stop loss order")

					// sell asap - we have no stops in place
					var order cb.Order
					if order, err = CreateOrder(pattern, pattern.NewMarketExitOrder()); err != nil {
						log.Error().
							Err(err).
							Uint("userID", pattern.UserID).
							Str("productID", pattern.Currency()).
							Float64("price", price).
							Float64("goal", goal).
							Msg("error while creating market exit order after failing to create stop loss order during sell")
						return err // this is bad, and super rare ... we need to start over
					}

					log.Info().
						Uint("userID", pattern.UserID).
						Str("productID", pattern.Currency()).
						Float64("goal", goal).
						Float64("exit", util.StringToFloat64(order.ExecutedValue)/util.StringToFloat64(order.Size)).
						Msg("sold")

					return nil // we have a nice stop loss in place, no need to try and sell this fill again
				}
			}
		} // else, onto the next rate and hopefully another higher price, woot!
	}
}
