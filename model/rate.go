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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"nuchal-api/db"
	"time"
)

type Rate struct {
	UnixSecond int64  `json:"unix_second" gorm:"primarykey"`
	ProductID  string `json:"product_id" gorm:"primarykey"`

	CreatedAt int64          `json:"created_at" gorm:"autoCreateTime:nano"`
	UpdatedAt int64          `json:"updated_at" gorm:"autoUpdateTime:nano"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"autoDeleteTime:nano;index"`

	Low     float64 `json:"low"`
	High    float64 `json:"high"`
	Open    float64 `json:"open"`
	Close   float64 `json:"close"`
	Volume  float64 `json:"volume"`
	Product Product `json:"product"`
}

func NewRate(productID string, rate cb.HistoricRate) Rate {
	return Rate{
		ProductID:  productID,
		UnixSecond: rate.Time.Unix(),
		Low:        rate.Low,
		High:       rate.High,
		Open:       rate.Open,
		Close:      rate.Close,
		Volume:     rate.Volume,
	}
}

func init() {
	db.Migrate(&Rate{})
}

func (r Rate) log() *zerolog.Logger {
	logger := log.
		With().
		Str("productID", r.ProductID).
		Float64("open", r.Open).
		Float64("high", r.High).
		Float64("low", r.Low).
		Float64("close", r.Close).
		Logger()
	return &logger
}

func (r *Rate) IsDown() bool {
	return r.Open > r.Close
}

func (r *Rate) IsUp() bool {
	return !r.IsDown()
}

func (r *Rate) IsInit() bool {
	return r != nil && r != (&Rate{})
}

func (r *Rate) avg() float64 {
	return (r.Open + r.High + r.Low + r.Close) / 4
}

func (r *Rate) Time() time.Time {
	return time.Unix(r.UnixSecond, 0)
}

func (r *Rate) data() []interface{} {
	return []interface{}{r.Time().UnixMilli(), r.Open, r.High, r.Low, r.Close, r.Volume}
}

func FindFirstRateByProductIDInTimeDescOrder(productID string, r *Rate) {
	db.Resolve().
		Preload("Product").
		Where("product_id = ?", productID).
		Order("unix_second desc").
		First(r)
}

func FindFirstRateByProductIDAndLessThanTimeInTimeDescOrder(productID string, before int64, r *Rate) {
	db.Resolve().
		Preload("Product").
		Where("product_id = ?", productID).
		Where("unix_second < ?", before).
		Order("unix_second desc").
		First(r)
}

func FindRates(productID string, alpha, omega int64) []Rate {
	var rates []Rate

	log.Trace().
		Str("productID", productID).
		Int64("alpha", alpha).
		Int64("omega", omega).
		Msg("find rates")

	db.Resolve().
		Preload("Product").
		Where("product_id = ?", productID).
		Where("unix_second BETWEEN ? AND ?", alpha, omega).
		Order("unix_second asc").
		Find(&rates)

	log.Trace().
		Str("productID", productID).
		Int64("alpha", alpha).
		Int64("omega", omega).
		Int("rates", len(rates)).
		Msg("find rates")

	return rates
}

// GetRates is the primary method for getting rates.
func GetRates(userID uint, productID string, alpha, omega int64) ([]Rate, error) {

	rates := FindRates(productID, alpha, omega)

	from := time.Unix(alpha, 0)
	to := time.Unix(omega, 0)

	if len(rates) > 0 {
		from = rates[len(rates)-1].Time().UTC()
	}

	out, err := GetHistoricRates(userID, productID, from, to)
	if err != nil {
		return nil, err
	}

	for _, rate := range out {
		newRate := NewRate(productID, rate)
		tx := db.Resolve().Create(&newRate)
		if tx.Error != nil {
			db.Resolve().Save(&newRate)
		}
	}

	return FindRates(productID, alpha, omega), nil
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
		end = start.Add(time.Hour * 4)
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

			rate := NewRate(productID, cb.HistoricRate{
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
