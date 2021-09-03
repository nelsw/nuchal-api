/*
 *
 * Copyright © 2021 Connor Van Elswyk ConnorVanElswyk@gmail.com
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/db"
	"sort"
	"time"
)

type Rate struct {
	Unix      int64  `json:"unix" gorm:"primaryKey"`
	ProductID string `json:"product_id" gorm:"primaryKey"`
	cb.HistoricRate
}

func NewRate(productID string, historicRate cb.HistoricRate) *Rate {
	rate := new(Rate)
	rate.Unix = historicRate.Time.UnixNano()
	rate.ProductID = productID
	rate.HistoricRate = historicRate
	return rate
}

func (v *Rate) IsDown() bool {
	return v.Open > v.Close
}

func (v *Rate) IsUp() bool {
	return !v.IsDown()
}

func (v *Rate) IsInit() bool {
	return v != nil && v != (&Rate{})
}

func (v *Rate) Time() time.Time {
	return time.Unix(0, v.Unix)
}

func (v *Rate) Label() string {
	return time.Unix(0, v.Unix).Format(time.Kitchen)
}

func (v *Rate) Data() []float64 {
	return []float64{v.Open, v.High, v.Low, v.Close}
}

func GetAllRatesBetween(userID uint, productID string, alpha, omega int64) []Rate {

	rates := GetOldRatesBetween(productID, alpha, omega)

	var from time.Time
	size := len(rates)
	if size > 0 {
		from = rates[size-1].Time()
	} else {
		from = time.Unix(alpha, 0)
	}

	newRates := GetNewRatesFrom(userID, productID, from)
	rates = append(newRates, rates...)

	sort.SliceStable(rates, func(i, j int) bool {
		return rates[i].Time().Before(rates[j].Time())
	})

	return rates
}

func GetOldRatesBetween(productID string, alpha, omega int64) []Rate {

	var rates []Rate

	db.Resolve().
		Where("product_id = ?", productID).
		Where("unix BETWEEN ? AND ?", alpha, omega).
		Order("unix desc").
		Find(&rates)

	return rates
}

func GetNewRatesFrom(userID uint, productID string, alpha time.Time) []Rate {

	var rates []Rate

	out, err := GetHistoricRates(userID, productID, alpha, time.Now())
	if err != nil {
		log.Err(err).Send()
		return rates
	}

	for _, rate := range out {
		modelRate := *NewRate(productID, rate)
		db.Resolve().Create(&modelRate)
		rates = append(rates, modelRate)
	}

	return rates
}

func GetHistoricRates(userID uint, productID string, alpha, omega time.Time) ([]cb.HistoricRate, error) {

	var rates []cb.HistoricRate

	params := rateParams(alpha, omega)

	u := FindUserByID(userID)

	for _, params := range params {
		out, err := u.Client().GetHistoricRates(productID, params)
		if err != nil {
			return nil, err
		}
		rates = append(rates, out...)
	}

	return rates, nil
}

func rateParams(alpha, omega time.Time) []cb.GetHistoricRatesParams {
	start := alpha
	end := start.Add(time.Hour * 4)
	var results []cb.GetHistoricRatesParams
	for i := 0; i < 24; i += 4 {
		results = append(results, cb.GetHistoricRatesParams{start, end, 60})
		start = end
		end = end.Add(time.Hour * 4)
		if end.After(omega) {
			break
		}
	}
	return results
}
