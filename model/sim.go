package model

import (
	"fmt"

	"nuchal-api/util"
	"time"
)

type Sim struct {
	Chart    `json:"chart"`
	Analysis `json:"analysis"`
	Pattern  `json:"pattern"`
}

type Analysis struct {
	Investment string    `json:"investment"`
	Return     string    `json:"return"`
	Percent    string    `json:"percent"`
	Summaries  []Summary `json:"summaries"`
}

type Summary struct {
	TradeNumber int    `json:"trade_number"`
	BuyPrice    string `json:"buy_price"`
	BuyTime     string `json:"buy_time"`
	SellPrice   string `json:"sell_price"`
	SellTime    string `json:"sell_time"`
	Net         string `json:"net"`
	Gross       string `json:"gross"`
	Profit      string `json:"profit"`
	Percent     string `json:"percent"`
}

type TradeType string

const (

	// lossType when the trade reaches a stop loss limit defined by the pattern
	lossType TradeType = "loss"

	// evenType when the trade reaches a break even limit defined by nuchal
	evenType = "even"

	// goalType when the sellOrder  meets or exceeds the goal price
	goalType = "goal"

	// humpType when the trade exceeds the price thanks to nuchal climb algo
	humpType = "hump"

	// downType when the trade doesn't finish and ends down
	downType = "down"

	// upType when the trade doesn't finish and ends down
	upType = "up"
)

type Trade struct {
	Index   int       `json:"index"`
	Buy     Rate      `json:"buyOrder"`
	Sell    Rate      `json:"sellOrder"`
	Type    TradeType `json:"trade_type"`
	Pattern Pattern   `json:"-"`
	Maker   float64   `json:"maker"`
	Taker   float64   `json:"taker"`
}

func newTrade(index int, pattern Pattern) *Trade {
	trade := new(Trade)
	trade.Index = index
	trade.Pattern = pattern
	user := pattern.user()
	trade.Maker = user.Maker
	trade.Taker = user.Taker
	return trade
}

func (t *Trade) in() float64 {
	return t.Buy.Open
}

func (t *Trade) out() float64 {
	if t.Type == lossType {
		return t.loss()
	} else if t.Type == evenType {
		return t.even()
	} else if t.Type == goalType {
		return t.goal()
	} else if t.Type == humpType {
		return t.Sell.Open
	} else if t.Type == downType || t.Type == upType {
		return t.Sell.Close
	} else {
		return 0.0
	}
}

func (t *Trade) color() string {
	if t.Type == lossType {
		return "#FF5722"
	} else if t.Type == evenType {
		return "#795548"
	} else if t.Type == goalType {
		return "#8BC34A"
	} else if t.Type == humpType {
		return "#4CAF50"
	} else if t.Type == downType {
		return "#FF9800"
	} else if t.Type == upType {
		return "#CDDC39"
	} else {
		return "#757575"
	}
}

func (t *Trade) text() string {
	if t.Type == lossType {
		return fmt.Sprintf("%d - %s", t.Index, `💩`)
	} else if t.Type == evenType {
		return fmt.Sprintf("%d - %s", t.Index, `👍🏻`)
	} else if t.Type == goalType {
		return fmt.Sprintf("%d - %s", t.Index, `🎯`)
	} else if t.Type == humpType {
		return fmt.Sprintf("%d - %s", t.Index, `🔥`)
	} else if t.Type == downType {
		return fmt.Sprintf("%d - %s", t.Index, `📉`)
	} else if t.Type == upType {
		return fmt.Sprintf("%d - %s", t.Index, `📈`)
	} else {
		return fmt.Sprintf("%d", t.Index)
	}
}

func (t *Trade) goal() float64 {
	return t.Pattern.GoalPrice(t.in())
}

func (t *Trade) loss() float64 {
	return t.Pattern.LossPrice(t.in())
}

func (t *Trade) even() float64 {
	return t.in() + (t.in() * (t.Taker + t.Maker))
}

func (t *Trade) investment() float64 {
	return t.in() * t.Pattern.Size
}

func (t *Trade) net() float64 {
	return t.out() - t.in()
}

func (t *Trade) gross() float64 {
	return t.net() * t.Pattern.Size
}

func (t *Trade) profit() float64 {
	return t.gross() - (t.gross() * (t.Taker + t.Maker))
}

func (t *Trade) percent() float64 {
	return ((t.out() / t.in()) - 1) * 100
}

func (t *Trade) orderData() [][]interface{} {
	return [][]interface{}{
		{t.Buy.Time().UnixMilli(), 1, t.in()},
		{t.Sell.Time().UnixMilli(), 0, t.out()},
	}
}

func (t *Trade) splitData() [][]interface{} {
	i := float64(t.Index%38) * 25.0 / 1000
	return [][]interface{}{
		{t.Buy.Time().UnixMilli(), fmt.Sprintf("%d", t.Index), 0, "#757575", i},
		{t.Sell.Time().UnixMilli(), t.text(), 1, t.color(), i},
	}
}

func (t *Trade) summary() Summary {
	return Summary{
		TradeNumber: t.Index,
		BuyTime:     t.Buy.Time().UTC().Format(time.Stamp),
		SellTime:    t.Sell.Time().UTC().Format(time.Stamp),
		BuyPrice:    fmt.Sprintf("$%f", t.in()),
		SellPrice:   fmt.Sprintf("$%f", t.out()),
		Net:         fmt.Sprintf("$%.2f", t.net()),
		Gross:       fmt.Sprintf("$%.2f", t.gross()),
		Profit:      fmt.Sprintf("$%.2f", t.profit()),
		Percent:     fmt.Sprintf("%.2f", t.percent()) + "%",
	}
}

func NewSim(patternID uint, alpha, omega int64) (sim Sim, err error) {

	pattern := FindPatternByID(patternID)

	var rates []Rate
	if rates, err = GetRates(pattern.UserID, pattern.ProductID, alpha, omega); err != nil {
		return
	}

	var candles, orders, splits [][]interface{}
	for _, rate := range rates {
		candles = append(candles, rate.data())
	}

	var summaries []Summary
	var inv, roi float64
	var then, that Rate

	for i, this := range rates {

		index := len(summaries)
		if pattern.Bound == "Buys" && index == pattern.Break {
			break
		}

		if pattern.MatchesTweezerBottomPattern(then, that, this) {
			index++
			trade := newTrade(index, pattern)
			trade.em(rates[i+1:])
			roi += trade.profit()
			inv += trade.investment()
			orders = append(orders, trade.orderData()...)
			splits = append(splits, trade.splitData()...)
			summaries = append(summaries, trade.summary())
		}

		then = that
		that = this
	}

	sim.Pattern = pattern
	sim.Chart = Chart{
		Layer{candleLayer, "", candles, Settings{}},
		[]Layer{
			{orderLayer, "Trades", orders, Settings{Legend: false, ZIndex: 5}},
			{splitterLayer, "Splits", splits, Settings{Legend: false, ZIndex: 10, LineWidth: 10}},
		},
	}
	sim.Analysis = Analysis{
		util.FloatToUsd(inv),
		util.FloatToUsd(roi),
		util.FloatToDecimal(roi / inv * 100),
		summaries,
	}
	return
}

func (t *Trade) em(rates []Rate) {

	for i, rate := range rates {

		if i == 0 {
			t.Buy = rate
		}

		// if the low meets or exceeds our loss limit ...
		if rate.Low <= t.loss() || rate.Low < t.out() {

			// ok, not the worst thing in the world, maybe a loss order already sold this for us
			if t.out() == 0.0 {
				// nope, we never established a loss order for this chart, we took a bath
				t.Type = lossType
				t.Sell = rate
			}
			break
		}

		// else if the high meets or exceeds our gain limit ...
		if rate.High >= t.goal() {

			// is this the first time this has happened?
			if t.out() == 0.0 {
				// yes, so now we have a loss (limit) buyOrder order placed, continue on.
				t.Type = goalType
				t.Sell = rate
			}

			// if we're here, a sellOrder limit has been set,
			// since we only set orders at the open
			// check if we placed a new order
			if t.out() < rate.Open {
				t.Type = humpType
				t.Sell = rate
			}
			continue
		}

		// the open of this rate is lower than our limit order
		if t.out() > 0.0 && t.out() > rate.Open {
			t.Sell = rate
			break
		}

		// if this is the first rate, give nuchal time orderType
		if i == 0 {
			continue
		}

		// if trading after 12 hours, try to break even
		if i > 60*12 && rate.High >= t.even() {
			t.Type = evenType
			t.Sell = rate
			break
		}

		// we're holding the orderType
		if i == len(rates)-1 {
			t.Sell = rate
			if rate.Close >= t.even() {
				t.Type = upType
			} else {
				t.Type = downType
			}
		}
	}
}
