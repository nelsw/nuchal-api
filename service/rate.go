package service

import (
	"github.com/rs/zerolog/log"
	"nuchal-api/db"
	"nuchal-api/model"
	"time"
)

func SaveTodayRatesFor(userID uint, productID string) error {

	if err := model.InitProducts(userID); err != nil {
		log.Err(err).Send()
		return err
	}

	log.Trace().Str("productID", productID).Send()

	omega := time.Now()
	alpha := time.Date(omega.Year(), omega.Month(), omega.Day(), 0, 0, 0, 0, time.UTC)

	history, err := model.GetHistoricRates(userID, productID, alpha, omega)
	if err != nil {
		log.Err(err).Send()
		return err
	}

	log.Trace().Int("history", len(history)).Send()

	for _, h := range history {
		rate := model.NewRate(productID, h)
		db.Resolve().Create(&rate)
	}

	return nil
}

func SaveAllNewRates(userID uint) error {

	if err := model.InitProducts(userID); err != nil {
		log.Err(err).Send()
		return err
	}

	for _, productID := range model.ProductIDs {

		log.Trace().Str("productID", productID).Send()

		t := getLastRateTime(productID)

		log.Trace().Time("last rate time", t).Send()

		history, err := model.GetHistoricRates(userID, productID, t, time.Now())
		if err != nil {
			log.Err(err).Send()
			return err
		}

		log.Trace().Int("history", len(history)).Send()

		for _, h := range history {
			rate := model.NewRate(productID, h)
			db.Resolve().Create(&rate)
		}
	}

	return nil
}

func getLastRateTime(productID string) time.Time {
	var rate model.Rate
	db.Resolve().
		Where("product_id = ?", productID).
		Order("unix desc").
		First(&rate)
	return rate.Time()
}
