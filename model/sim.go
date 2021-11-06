package model

import (
	"fmt"

	"nuchal-api/util"
	"strconv"
	"time"
)

type sideType int

const (
	left sideType = iota
	right
)

type tradeType int

const (
	sell tradeType = iota
	buy
)

type DataType string

const (
	trade    DataType = "Trades"
	candle            = "Candles"
	splitter          = "Splitters"
)

type Analysis struct {
	Investment string    `json:"investment"`
	Return     string    `json:"return"`
	Percent    string    `json:"percent"`
	Summaries  []Summary `json:"summaries"`
}

type Summary struct {
	TradeNumber              int    `json:"trade_number"`
	BuyPrice                 string `json:"buy_price"`
	BuyTime                  string `json:"buy_time"`
	SellPrice                string `json:"sell_price"`
	SellTime                 string `json:"sell_time"`
	Net                      string `json:"net"`
	Percent                  string `json:"percent"`
	Gross                    string `json:"gross"`
	Profit                   string `json:"profit"`
	investment, roi, percent float64
}

type Response struct {
	Chart    Result   `json:"chart"`
	OnChart  []Result `json:"onchart"`
	Analysis Analysis `json:"analysis"`
}

type Result struct {
	DataType `json:"type"`
	Settings `json:"settings"`
	Name     string          `json:"name"`
	Data     [][]interface{} `json:"data"`
}

type Settings struct {
	Legend bool `json:"legend"`
	ZIndex int  `json:"z-index"`
}

type action struct {
	tradeType
	time.Time
	price float64
}

func (a action) toData() []interface{} {
	return []interface{}{a.UnixMilli(), a.tradeType, a.price}
}

type split struct {
	sideType
	time.Time
	place       float64
	text, color string
}

func (s split) toData() []interface{} {
	return []interface{}{s.UnixMilli(), s.text, s.sideType, s.color, s.place}
}

func NewProductSim(userID uint, productID uint, alpha, omega int64) Response {
	pid := FindProductByID(productID).PID()
	chart := Result{candle, Settings{}, pid, nil}

	rates := GetAllRatesBetween(userID, productID, alpha, omega)
	for _, rate := range rates {
		chart.Data = append(chart.Data, rate.OHLCV())
	}
	splits := Result{splitter, Settings{false, 10}, "Splits", [][]interface{}{}}
	trades := Result{trade, Settings{false, 5}, "Trades", [][]interface{}{}}
	return Response{Chart: chart, OnChart: []Result{trades, splits}}
}

func NewPatternSim(patternID uint, alpha, omega int64) Response {
	return newSwim(FindPatternByID(patternID), alpha, omega)
}

func newSwim(pattern Pattern, alpha, omega int64) Response {

	splits := Result{splitter, Settings{false, 10}, "Splits", nil}
	trades := Result{trade, Settings{false, 5}, "Trades", nil}
	chart := Result{candle, Settings{}, pattern.Currency(), nil}

	rates := GetAllRatesBetween(pattern.UserID, pattern.ProductID, alpha, omega)
	for _, rate := range rates {
		chart.Data = append(chart.Data, rate.OHLCV())
	}
	user := FindUserByID(pattern.UserID)

	var summaries []Summary
	var then, that Rate
	for i, this := range rates {

		trx := len(summaries) + 1

		if pattern.MatchesTweezerBottomPattern(then, that, this) {
			summary := Summary{
				TradeNumber: trx,
			}
			handleOpportunity(&summary, &trades, &splits, user, pattern, rates[i+1:])
			summaries = append(summaries, summary)
		}

		if pattern.Bound == "Buys" && trx-1 == pattern.Break {
			break
		}

		then = that
		that = this
	}

	var invest, roi float64
	for _, summary := range summaries {
		invest += summary.investment
		roi += summary.roi
	}

	return Response{
		chart,
		[]Result{trades, splits},
		Analysis{
			util.FloatToUsd(invest),
			util.FloatToUsd(roi),
			util.FloatToDecimal(roi / invest * 100),
			summaries,
		},
	}
}

func handleOpportunity(summary *Summary, trades, splits *Result, user User, pattern Pattern, rates []Rate) {

	var on, off split
	var in, out action

	var goal, loss, even float64

	trx := strconv.Itoa(summary.TradeNumber)

	for i, rate := range rates {

		if i == 0 {
			act(&in, rate.Time(), rate.Open, buy)
			spl(&on, rate.Time(), trx, "#777", 0.75, left)

			goal = pattern.GoalPrice(in.price)
			loss = pattern.LossPrice(in.price)
			even = (in.price * user.Taker) + (in.price * user.Maker) + in.price
		}

		// if the low meets or exceeds our loss limit ...
		if rate.Low <= loss || rate.Low < out.price {

			// ok, not the worst thing in the world, maybe a loss order already sold this for us
			if out.price == 0.0 {
				// nope, we never established a loss order for this chart, we took a bath
				act(&out, rate.Time(), loss, sell)
				spl(&off, rate.Time(), trx, "#f4c20d", 0.75, right)
			}
			break
		}

		// else if the high meets or exceeds our gain limit ...
		if rate.High >= goal {

			// is this the first time this has happened?
			if out.price == 0.0 {
				// yes, so now we have a loss (limit) buy order placed, continue on.
				act(&out, rate.Time(), goal, sell)
				spl(&off, rate.Time(), trx, "#f4c20d", 0.75, right)
			}

			// if we're here, a sell limit has been set,
			// since we only set orders at the open
			// check if we placed a new order
			if out.price < rate.Open {
				act(&out, rate.Time(), rate.Open, sell)
				spl(&off, rate.Time(), trx, "#f4c20d", 0.75, right)
			}
			continue
		}

		// the open of this rate is lower than our limit order
		if out.price > 0.0 && out.price > rate.Open {
			break
		}

		// if this is the first rate, give nuchal time tradeType
		if i == 0 {
			continue
		}

		// if we have yet to goal this product through a limit order and we can break even, goal it
		if i > 60*12 && rate.High >= even {
			act(&out, rate.Time(), even, sell)
			spl(&off, rate.Time(), trx, "#f4c20d", 0.75, right)
			break
		}

		// we're holding the tradeType
		if i == len(rates)-1 {
			act(&out, rate.Time(), rate.Close, sell)
			spl(&off, rate.Time(), trx, "#f4c20d", 0.75, right)
		}
	}

	trades.Data = append(trades.Data, in.toData(), out.toData())
	splits.Data = append(splits.Data, on.toData(), off.toData())

	summary.investment = in.price * pattern.Size
	summary.BuyPrice = fmt.Sprintf("$%f", in.price)
	summary.BuyTime = time.Unix(in.Unix(), 0).UTC().Format(time.Stamp)
	summary.SellPrice = fmt.Sprintf("$%f", out.price)
	summary.SellTime = time.Unix(out.Unix(), 0).UTC().Format(time.Stamp)
	summary.Percent = fmt.Sprintf("%.2f", ((out.price/in.price)-1)*100) + "%"

	net := out.price - in.price
	gross := net * pattern.Size
	profit := gross - (gross * (user.Maker + user.Taker))

	summary.roi = profit
	summary.percent = profit / summary.investment * 100
	summary.Net = fmt.Sprintf("$%.2f", net)
	summary.Gross = fmt.Sprintf("$%.2f", gross)
	summary.Profit = fmt.Sprintf("$%.2f", profit)
}

func act(a *action, ts time.Time, price float64, tt tradeType) {
	a.Time = ts
	a.price = price
	a.tradeType = tt
}

func spl(s *split, ts time.Time, text, color string, place float64, st sideType) {
	s.Time = ts
	s.text = text
	s.color = color
	s.place = place
	s.sideType = st
}
