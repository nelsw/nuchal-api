package view

import (
	"fmt"
	"nuchal-api/model"
	"nuchal-api/util"
	"time"
)

const (
	Enter = `#651FFF`
	Won   = `#69F0AE`
	Lost  = `#F50057`
	Trade = `#4E342E`
)

var (
	detail  = Detail{"candlestick", 500}
	tooltip = Tooltip{true, "dark"}
	xaxis   = XAxis{tooltip, "datetime"}
	yaxis   = YAxis{tooltip}
)

type Chart struct {
	Series     []Series `json:"series"`
	Options    `json:"chartOptions"`
	Simulation `json:"simulation"`
}

type Series struct {
	Name string       `json:"name"`
	Data []model.Data `json:"data"`
}

type Options struct {
	Detail      `json:"chart"`
	Title       `json:"title"`
	Tooltip     `json:"tooltip"`
	XAxis       `json:"xaxis"`
	YAxis       `json:"yaxis"`
	Annotations `json:"annotations"`
}

type Detail struct {
	Type   string `json:"type"`
	Height int    `json:"height"`
}

type Title struct {
	Text  string `json:"text"`
	Align string `json:"align"`
}

type Tooltip struct {
	Enabled bool   `json:"enabled"`
	Theme   string `json:"theme"`
}

type XAxis struct {
	Tooltip `json:"tooltip"`
	Type    string `json:"type"`
}

type YAxis struct {
	Tooltip `json:"tooltip"`
}

type Annotations struct {
	XAxis []Annotation `json:"xaxis"`
}

type Annotation struct {
	X     int64  `json:"x"`
	Color string `json:"borderColor"`
	Label `json:"label"`
}

type Label struct {
	BorderColor string `json:"borderColor"`
	Orientation string `json:"orientation"`
	OffsetY     int    `json:"offsetY"`
	Text        string `json:"text"`
	Style       `json:"style"`
}

type Style struct {
	FontSize   string `json:"fontSize"`
	Color      string `json:"color"`
	Background string `json:"background"`
}

type Simulation struct {

	// Gain is the gross amount won, fees not included.
	Gain float64 `json:"gain"`

	// Loss is the gross amount lost, fees not included.
	Loss float64 `json:"loss"`

	// Rise are the trades we're holding which are currently greater than the buy.
	Rise float64 `json:"rise"`

	// Fall are the trades we're holding which are currently less than the buy.
	Fall float64 `json:"fall"`

	// Stake is the amount we invested.
	Stake float64 `json:"stake"`

	// Take is the result of summing the total with closed holds
	Take float64 `json:"take"`

	// Net profit or loss, fees included. Does not include held products.
	Net float64 `json:"net"`

	// Total is the net less maker and taker fees.
	Total float64 `json:"total"`

	// Transactions are the trades that comprise the simulation.
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {

	// Alpha is the time of the buy.
	Alpha int64 `json:"alpha"`

	// Omega is the time of the sell.
	Omega int64 `json:"omega"`

	// Buy is the price which we purchase the product.
	Buy float64 `json:"buy"`

	// Sell is the price which we want to sell the product.
	Sell float64 `json:"sell"`

	// Sold is the price which we sold the product.
	Sold float64 `json:"sold"`

	// Hold is the last known price of the trade, it's still active.
	Hold float64 `json:"hold"`

	// Goal is the price which we want to sell that product including fees.
	Goal float64 `json:"goal"`

	// Stop is the price which we sell no matter what.
	Stop float64 `json:"stop"`

	// Entry is the buy price plus maker fees.
	Entry float64 `json:"entry"`

	// Exit is sell price plus taker fees.
	Exit float64 `json:"exit"`

	// Result is the net less maker and taker fees.
	Result float64 `json:"result"`

	AveragePrices []float64 `json:"average_prices"`
}

var offset int

func NewAnnotation(x int64, color, text string) Annotation {
	if offset < 300 {
		offset += 12
	} else {
		offset = 0
	}

	return Annotation{
		X:     x,
		Color: color,
		Label: Label{
			BorderColor: color,
			Orientation: "horizontal",
			OffsetY:     offset,
			Text:        text,
			Style: Style{
				"12px",
				"#000",
				color,
			},
		},
	}
}

// todo - include size in calcs
func NewChartData(userID uint, productID string, alpha, omega int64) Chart {

	rates := model.GetAllRatesBetween(userID, productID, alpha, omega)

	pattern := model.GetPattern(userID, productID)

	user := model.FindUserByID(userID)

	var gain, loss, holdUp, holdDown, stake float64

	var annotations []Annotation
	var transactions []Transaction

	var then, that model.Rate
	for i, this := range rates {

		if pattern.Bound == "Buys" && len(transactions) == pattern.Break {
			break
		}

		if pattern.MatchesTweezerBottomPattern(then, that, this) {

			var averagePrices []float64

			var buy, sell, sold, goal, entry, stop float64
			var lastRate model.Rate

			for j, rate := range rates[i:] {

				averagePrices = append(averagePrices, rate.AveragePrice())

				if j == 0 {
					sold = -1.0
					buy = rate.Open
					stake += buy * pattern.Size
					sell = pattern.GoalPrice(buy)
					stop = pattern.LossPrice(buy)
					goal = sell + (sell * user.Maker)
					entry = buy + (buy * user.Taker)
				}

				lastRate = rate

				// if the low meets or exceeds our loss limit ...
				if rate.Low < stop {

					// ok, not the worst thing in the world, maybe a stop order already sold this for us
					if sold < 0 {
						// nope, we never established a stop order for this chart, we took a bath
						sold = stop
					}
					break
				}

				// else if the high meets or exceeds our gain limit ...
				if rate.High >= goal {

					// is this the first time this has happened?
					if sold < 0 {
						// great, we have a stop (limit) buy order placed, continue on.
						sold = sell
					}

					// now if the rate closes less than our sold, the buy order would have been triggered.
					if rate.Close < goal {
						break
					}

					// else we're trending up, ride the wave.
					if rate.Close >= goal {
						sold = rate.Close
					}
				}

				// else the low and highs of this rate do not exceed either respective limit
				// we must now navigate each rate and attempt to sell at profit
				// and avoid the ether
				if j == 0 {
					continue
				}

				// if we have yet to sell this product
				if sold < 0 &&
					// and we've been holding for some time
					rate.Time().Sub(this.Time()) > time.Hour*12 &&
					// and we can break even ...
					rate.High >= entry+(entry*user.Maker) {
					// sell it
					sold = entry
					break
				}
			}

			alpha := this.Time().UTC().Unix()
			omega := lastRate.Time().UTC().Unix()

			var icon, sale string
			var hold, exit, result float64

			entry *= pattern.Size

			if sold < 0 {
				icon = Trade
				hold = lastRate.Close
				exit = (lastRate.Close - (lastRate.Close * user.Maker)) * pattern.Size
				result = exit - entry
				sale = util.FloatToUsd(hold)
				if hold >= buy {
					holdUp += hold
				} else {
					holdDown += hold
				}
			} else {
				exit = (sold - (sold * user.Maker)) * pattern.Size
				result = exit - entry
				sale = util.FloatToUsd(sold)
				if result > -0 {
					gain += result
					icon = Won
				} else {
					gain -= result
					icon = Lost

				}
			}

			annotations = append(annotations, NewAnnotation(alpha, Enter, util.FloatToUsd(buy)))
			annotations = append(annotations, NewAnnotation(omega, icon, sale))
			transactions = append(transactions, Transaction{
				Alpha:         alpha,
				Omega:         omega,
				Buy:           buy,
				Sell:          sell,
				Stop:          stop,
				Sold:          sold,
				Hold:          hold,
				Goal:          goal,
				Entry:         entry,
				Exit:          exit,
				Result:        result,
				AveragePrices: averagePrices,
			})
		}
		then = that
		that = this
	}

	title := fmt.Sprintf("%s=%f ... %s=%f ... %s=%f ... %s=%f",
		util.Won, gain,
		util.Lost, loss,
		util.TradingUp, holdUp,
		util.TradingDown, holdDown,
	)

	var data []model.Data
	for _, rate := range rates {
		data = append(data, rate.Data())
	}

	return Chart{
		Simulation: Simulation{
			Gain:         gain,
			Loss:         loss,
			Rise:         holdUp,
			Fall:         holdDown,
			Stake:        stake,
			Transactions: transactions,
		},
		Series: []Series{{"candles", data}},
		Options: Options{
			Detail:      detail,
			Title:       Title{title, "left"},
			Tooltip:     tooltip,
			XAxis:       xaxis,
			YAxis:       yaxis,
			Annotations: Annotations{annotations},
		},
	}
}
