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
	Fees       string    `json:"fees"`
	Return     string    `json:"return"`
	Percent    string    `json:"percent"`
	Summaries  []Summary `json:"summaries"`
}

type Summary struct {
	TradeNumber int64  `json:"trade_number"`
	BuyPrice    string `json:"buy_price"`
	BuyTime     string `json:"buy_time"`
	SellPrice   string `json:"sell_price"`
	SellTime    string `json:"sell_time"`
	Net         string `json:"net"`
	Gross       string `json:"gross"`
	Profit      string `json:"profit"`
	Percent     string `json:"percent"`
	Color       string `json:"color"`
	Emoji       string `json:"emoji"`
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

type MockTrade struct {
	Index   int64     `json:"index"`
	Buy     Rate      `json:"buyOrder"`
	Sell    Rate      `json:"sellOrder"`
	Type    TradeType `json:"trade_type"`
	Pattern Pattern   `json:"-"`
	Maker   float64   `json:"maker"`
	Taker   float64   `json:"taker"`
}

func newTrade(index int64, pattern Pattern) *MockTrade {
	trade := new(MockTrade)
	trade.Index = index
	trade.Pattern = pattern
	trade.Maker = pattern.User.Maker
	trade.Taker = pattern.User.Taker
	return trade
}

func (t *MockTrade) in() float64 {
	return t.Buy.Open
}

func (t *MockTrade) entry() float64 {
	return t.in() + (t.in() * t.Pattern.User.Maker)
}

func (t *MockTrade) exit() float64 {
	return t.out() + (t.out() * t.Pattern.User.Taker)
}

func (t *MockTrade) out() float64 {
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

func (t *MockTrade) color() string {
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

func (t *MockTrade) text() string {
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

func (t *MockTrade) emoji() string {
	if t.Type == lossType {
		return `💩`
	} else if t.Type == evenType {
		return `👍🏻`
	} else if t.Type == goalType {
		return `🎯`
	} else if t.Type == humpType {
		return `🔥`
	} else if t.Type == downType {
		return `📉`
	} else if t.Type == upType {
		return `📈`
	} else {
		return ``
	}
}

func (t *MockTrade) goal() float64 {
	return t.Pattern.GoalPrice(t.in())
}

func (t *MockTrade) loss() float64 {
	return t.Pattern.LossPrice(t.in())
}

func (t *MockTrade) even() float64 {
	entry := t.in() + (t.in() * t.Pattern.User.Maker)
	exit := entry + (entry * t.Pattern.User.Taker)
	return entry + exit
}

func (t *MockTrade) investment() float64 {
	return t.entry() * t.Pattern.Size
}

func (t *MockTrade) fees() float64 {
	return (t.in() * t.Pattern.User.Maker) + (t.out() * t.Pattern.User.Taker)
}

func (t *MockTrade) net() float64 {
	return t.out() - t.in()
}

func (t *MockTrade) gross() float64 {
	return t.exit() - t.entry()
}

func (t *MockTrade) profit() float64 {
	return (t.exit() - t.entry()) * t.Pattern.Size
}

func (t *MockTrade) percent() float64 {
	return t.profit() / t.investment() * 100
}

func (t *MockTrade) orderData() [][]interface{} {
	return [][]interface{}{
		{t.Buy.Time().UnixMilli(), 1, t.Pattern.Product.precise(t.in())},
		{t.Sell.Time().UnixMilli(), 0, t.Pattern.Product.precise(t.out())},
	}
}

func (t *MockTrade) splitData() [][]interface{} {
	i := float64(t.Index%38) * 25.0 / 1000
	return [][]interface{}{
		{t.Buy.Time().UnixMilli(), fmt.Sprintf("%d", t.Index), 0, "#757575", i},
		{t.Sell.Time().UnixMilli(), t.text(), 1, t.color(), i},
	}
}

func (t *MockTrade) summary() Summary {
	return Summary{
		TradeNumber: t.Index,
		BuyTime:     t.Buy.Time().UTC().Format(time.Stamp),
		SellTime:    t.Sell.Time().UTC().Format(time.Stamp),
		BuyPrice:    t.Pattern.Product.precise(t.in()),
		SellPrice:   t.Pattern.Product.precise(t.out()),
		Net:         t.Pattern.Product.precise(t.net()),
		Gross:       t.Pattern.Product.precise(t.gross()),
		Profit:      t.Pattern.Product.precise(t.profit()),
		Percent:     fmt.Sprintf("%.2f", t.percent()) + "%",
		Color:       t.color(),
		Emoji:       t.emoji(),
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
	var inv, roi, fee float64
	var then, that Rate

	for i, this := range rates {

		index := int64(len(summaries))
		if pattern.Bound == buyBound && index == pattern.Bind {
			break
		}

		if pattern.MatchesTweezerBottomPattern(then, that, this) {
			index++
			trade := newTrade(index, pattern)
			trade.em(rates[i+1:])
			fee += trade.fees()
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
		util.FloatToUsd(fee),
		util.FloatToUsd(roi),
		util.FloatToDecimal(roi / inv * 100),
		summaries,
	}
	return
}

func (t *MockTrade) em(rates []Rate) {

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
