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
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"nuchal-api/db"
	"sort"
	"time"
)

type Rate struct {
	gorm.Model
	ProductID       uint `json:"product_id"`
	cb.HistoricRate `gorm:"embedded"`
}

type Data struct {
	X int64     `json:"x"`
	Y []float64 `json:"y"`
}

func init() {
	db.Migrate(&Rate{})
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
	return v.HistoricRate.Time
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

func GetAllRatesBetween(userID uint, productID uint, alpha, omega int64) []Rate {

	var rates []Rate

	from := time.Unix(alpha, 0)
	to := time.Unix(omega, 0)

	db.Resolve().
		Where("product_id = ?", productID).
		Where("time BETWEEN ? AND ?", from, to).
		Order("time asc").
		Find(&rates)

	size := len(rates)

	if size > 0 {
		from = rates[size-1].Time()
	} else {
		from = time.Unix(alpha, 0)
	}

	pid := FindProductByID(productID).PID()

	out, err := GetHistoricRates(userID, pid, from, to)
	if err != nil {
		log.Err(err).Send()
		return rates
	}

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Time.Before(out[j].Time)
	})

	for _, rate := range out {
		rates = append(rates, Rate{HistoricRate: rate})
	}

	db.Resolve().Create(&rates)

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
			return nil, err
		}
		rates = append(rates, out...)
	}

	return rates, nil
}

func rateParams(alpha, omega time.Time) []cb.GetHistoricRatesParams {

	start := alpha
	end := alpha.Add(time.Hour * 4)

	var results []cb.GetHistoricRatesParams

	for i := 0.0; i < 24; i += 4 {
		results = append(results, cb.GetHistoricRatesParams{start, end, 60})
		start = end
		end = end.Add(time.Hour * 4)
		if start.After(omega) {
			break
		}
	}
	return results
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
				HistoricRate: cb.HistoricRate{
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
