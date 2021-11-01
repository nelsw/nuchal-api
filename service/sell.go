package service

import (
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/model"
	"nuchal-api/util"
)

func sell(wsConn *ws.Conn, entry float64, size string, pattern model.Pattern) error {

	goal := pattern.GoalPrice(entry)
	loss := pattern.LossPrice(entry)

	log.Info().
		Uint("userID", pattern.UserID).
		Str("productID", pattern.ProductID).
		Float64("entry", entry).
		Float64("goal", goal).
		Float64("loss", loss).
		Str("size", size).
		Msg("sell")

	var err error

	if _, err = CreateOrder(pattern.UserID, pattern.StopLossOrder(size, loss)); err != nil {
		log.Error().
			Err(err).
			Uint("userID", pattern.UserID).
			Str("productID", pattern.ProductID).
			Msg("error placing stop loss order for loss price price during sell")
		// just keep going, yolo my brolo
	}

	for {

		var price float64

		if price, err = getPrice(wsConn, pattern.ProductID); err != nil {
			log.Error().
				Err(err).
				Uint("userID", pattern.UserID).
				Str("productID", pattern.ProductID).
				Msg("error getting price during sell")

			// can't get price info so create a stop entry order in case the price reaches our goal
			if _, err = CreateOrder(pattern.UserID, pattern.NewStopEntryOrder(size, goal)); err != nil {
				log.Error().
					Err(err).
					Uint("userID", pattern.UserID).
					Str("productID", pattern.ProductID).
					Msg("error while creating stop entry order after price error during sell")
			}

			return err // can't proceed without price data, return an error and RUN IT AGAIN!
		}

		if price <= loss {
			log.Info().
				Uint("userID", pattern.UserID).
				Str("productID", pattern.ProductID).
				Float64("price", price).
				Float64("exit", loss).
				Msg("price <= loss, sold")
			return nil
		}

		if price < goal {
			continue
		}

		log.Info().
			Uint("userID", pattern.UserID).
			Str("productID", pattern.ProductID).
			Float64("price", price).
			Float64("goal", goal).
			Msg("price >= goal")

		var stopLoss cb.Order

		// place a stop loss at our goal and try to find a higher price
		if stopLoss, err = CreateOrder(pattern.UserID, pattern.StopLossOrder(size, goal)); err != nil {
			log.Error().
				Err(err).
				Uint("userID", pattern.UserID).
				Str("productID", pattern.ProductID).
				Float64("price", price).
				Float64("goal", goal).
				Msg("error while creating stop loss order after goal >= price during sell")

			// nvm, sell asap - hopefully the price is still higher than our goal
			var order cb.Order
			if order, err = CreateOrder(pattern.UserID, pattern.NewMarketExitOrder()); err != nil {
				log.Error().
					Err(err).
					Uint("userID", pattern.UserID).
					Str("productID", pattern.ProductID).
					Float64("price", price).
					Float64("goal", goal).
					Msg("error while creating market exit order after failing to create stop loss order during sell")

				return err // wow, we can't even place a market exit order, RUN IT AGAIN!
			}

			log.Info().
				Uint("userID", pattern.UserID).
				Str("productID", pattern.ProductID).
				Float64("goal", goal).
				Float64("exit", util.StringToFloat64(order.ExecutedValue)/util.StringToFloat64(order.Size)).
				Msg("price >= goal")

			return nil // we sold at market price so no point in returning an error and triggering another sell
		}

		for { // if we're here, we placed a stop loss order at our goal and now we try to find a better price

			var rate model.Rate

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

				if err = CancelOrder(pattern.UserID, stopLoss.ID); err != nil {
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
				if stopLoss, err = CreateOrder(pattern.UserID, pattern.StopLossOrder(size, goal)); err != nil {
					log.Error().
						Err(err).
						Uint("userID", pattern.UserID).
						Str("productID", pattern.ProductID).
						Msg("error while creating stop loss order")

					// sell asap - we have no stops in place
					var order cb.Order
					if order, err = CreateOrder(pattern.UserID, pattern.NewMarketExitOrder()); err != nil {
						log.Error().
							Err(err).
							Uint("userID", pattern.UserID).
							Str("productID", pattern.ProductID).
							Float64("price", price).
							Float64("goal", goal).
							Msg("error while creating market exit order after failing to create stop loss order during sell")
						return err // this is bad, and super rare ... we need to start over
					}

					log.Info().
						Uint("userID", pattern.UserID).
						Str("productID", pattern.ProductID).
						Float64("goal", goal).
						Float64("exit", util.StringToFloat64(order.ExecutedValue)/util.StringToFloat64(order.Size)).
						Msg("sold")

					return nil // we have a nice stop loss in place, no need to try and sell this fill again
				}
			}
		} // else, onto the next rate and hopefully another higher price, woot!
	}
}
