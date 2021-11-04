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
	"fmt"
	ws "github.com/gorilla/websocket"
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

type Data struct {
	X int64     `json:"x"`
	Y []float64 `json:"y"`
}

func init() {
	db.Migrate(&Rate{})
}

func NewRate(productID string, historicRate cb.HistoricRate) Rate {
	return Rate{
		Unix:         historicRate.Time.Unix(),
		ProductID:    productID,
		HistoricRate: historicRate,
	}
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
	return time.Unix(v.Unix, 0)
}

func (v *Rate) Data() Data {
	return Data{v.Time().UTC().Unix(), []float64{v.Open, v.High, v.Low, v.Close}}
}

func (v *Rate) OHLCV() []interface{} {
	return []interface{}{v.Time().UnixMilli(), v.Open, v.High, v.Low, v.Close, v.Volume}
}

func (v Rate) AveragePrice() float64 {
	return (v.Open + v.High + v.Low + v.Close) / 4
}

func (v Rate) Stamp() string {
	return v.Time().UTC().Format(time.Stamp)
}

func InitRates(userID uint) error {

	if err := InitProducts(userID); err != nil {
		return err
	}

	for _, product := range ProductArr {
		productID := product.BaseCurrency + "-USD"
		err := InitRate(userID, productID)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitRate(userID uint, productID string) error {

	log.Trace().
		Uint("userID", userID).
		Str("productID", productID).
		Msg("InitRate")

	alpha := time.Date(2021, 11, 2, 0, 0, 0, 0, time.UTC)
	omega := time.Date(2021, 11, 4, 0, 0, 0, 0, time.UTC).Add(time.Second * -1)

	rates := GetNewRatesFromTo(userID, productID, alpha, omega)

	log.Trace().
		Uint("userID", userID).
		Str("productID", productID).
		Time("alpha", alpha).
		Int("rates", len(rates)).
		Msg("InitRate")

	db.Resolve().CreateInBatches(&rates, 1000)

	return nil
}

func GetAllRatesBetween(userID uint, productID string, alpha, omega int64) []Rate {

	rates := FindRatesBetween(productID, alpha, omega)

	var from time.Time
	size := len(rates)
	if size > 0 {
		from = rates[size-1].Time()
	} else {
		from = time.Unix(alpha, 0)
	}

	newRates := GetNewRatesFrom(userID, productID, from)
	rates = append(rates, newRates...)

	sort.SliceStable(rates, func(i, j int) bool {
		return rates[i].Time().Before(rates[j].Time())
	})

	return rates
}

func FindRatesBetween(productID string, alpha, omega int64) []Rate {

	var rates []Rate

	db.Resolve().
		Where("product_id = ?", productID).
		Where("unix BETWEEN ? AND ?", alpha, omega).
		Order("unix asc").
		Find(&rates)

	return rates
}

func GetNewRatesFromTo(userID uint, productID string, alpha, omega time.Time) []Rate {

	log.Trace().
		Uint("userID", userID).
		Str("productID", productID).
		Time("alpha", alpha).
		Time("omega", omega).
		Msg("GetNewRatesFromTo")

	var rates []Rate

	out, err := GetHistoricRates(userID, productID, alpha, omega)
	if err != nil {
		log.Err(err).Send()
		return rates
	}

	for _, rate := range out {
		rates = append(rates, NewRate(productID, rate))
	}

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
		rates = append(rates, Rate{rate.Time.Unix(), productID, rate})
	}

	return rates
}

func GetHistoricRates(userID uint, productID string, alpha, omega time.Time) ([]cb.HistoricRate, error) {

	var rates []cb.HistoricRate

	params := rateParams(alpha, omega)

	u := FindUserByID(userID)

	for _, p := range params {
		out, err := u.Client().GetHistoricRates(productID, p)
		if err != nil {
			log.Err(err).Send()
			fmt.Println(err)
			return rates, err
		}
		rates = append(rates, out...)
	}

	return rates, nil
}

func rateParams(alpha, omega time.Time) []cb.GetHistoricRatesParams {

	start := omega.Add(time.Hour * 4 * -1)
	end := omega

	chunks := omega.Sub(alpha).Minutes()

	var results []cb.GetHistoricRatesParams

	for i := 0.0; i < chunks; i += 4 {
		results = append(results, cb.GetHistoricRatesParams{start, end, 60})
		end = start
		start = start.Add(time.Hour * 4 * -1)
		if start.Before(alpha) {
			break
		}
	}
	return results
}

func getLastRateTime(productID uint) time.Time {
	var rate Rate
	db.Resolve().
		Where("product_id = ?", productID).
		Order("unix desc").
		First(&rate)
	return rate.Time()
}

func rate(productID string) (Rate, error) {

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Error().Err(err).Msg("opening ws")
		return Rate{}, err
	}

	defer func(wsConn *ws.Conn) {
		if err = wsConn.Close(); err != nil {
			log.Error().Err(err).Msg("closing ws")
		}
	}(wsConn)

	if err = wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{productID}}},
	}); err != nil {
		log.Error().Err(err).Msg("writing ws")
		return Rate{}, err
	}

	return getRate(wsConn, productID)
}

func getRate(wsConn *ws.Conn, productID string) (Rate, error) {

	end := time.Now().Add(time.Minute)

	var low, high, open, volume float64
	for {

		price, err := getPrice(wsConn, productID)
		if err != nil {
			log.Error().Err(err).Str("productID", productID).Msg("price")
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

			log.Info().Str("productID", productID).Msg("rate")

			return rate, nil
		}
	}
}
