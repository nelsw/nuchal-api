package model

import (
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/db"
	"time"
)

func SaveTodayRatesFor(userID uint, productID string) error {

	if err := InitProducts(userID); err != nil {
		log.Err(err).Send()
		return err
	}

	log.Trace().Str("productID", productID).Send()

	omega := time.Now()
	alpha := time.Date(omega.Year(), omega.Month(), omega.Day(), 0, 0, 0, 0, time.UTC)

	history, err := GetHistoricRates(userID, productID, alpha, omega)
	if err != nil {
		log.Err(err).Send()
		return err
	}

	log.Trace().Int("history", len(history)).Send()

	for _, h := range history {
		rate := NewRate(productID, h)
		db.Resolve().Create(&rate)
	}

	return nil
}

func SaveAllNewRates(userID uint) error {

	if err := InitProducts(userID); err != nil {
		log.Err(err).Send()
		return err
	}

	for _, productID := range ProductIDs {

		log.Trace().Str("productID", productID).Send()

		t := getLastRateTime(productID)

		log.Trace().Time("last rate time", t).Send()

		history, err := GetHistoricRates(userID, productID, t, time.Now())
		if err != nil {
			log.Err(err).Send()
			return err
		}

		log.Trace().Int("history", len(history)).Send()

		for _, h := range history {
			rate := NewRate(productID, h)
			db.Resolve().Create(&rate)
		}
	}

	return nil
}

func getLastRateTime(productID string) time.Time {
	var rate Rate
	db.Resolve().
		Where("product_id = ?", productID).
		Order("unix desc").
		First(&rate)
	return rate.Time()
}

func getRate(wsConn *ws.Conn, productID string) (Rate, error) {

	end := time.Now().Add(time.Minute)

	var low, high, open, volume float64
	for {

		price, err := getPrice(wsConn, productID)
		if err != nil {
			log.Error().
				Err(err).
				Str("productID", productID).
				Msg("error getting price")
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

			rate := Rate{
				time.Now().UnixMilli(),
				productID,
				cb.HistoricRate{
					time.Now(),
					low,
					high,
					open,
					price,
					volume,
				},
			}

			log.Info().
				Str("productID", productID).
				Msg("got rate")

			return rate, nil
		}
	}
}
