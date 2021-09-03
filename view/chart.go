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
	Series  []Series `json:"series"`
	Options `json:"chartOptions"`
}

type Series struct {
	Name string `json:"name"`
	Data []Data `json:"data"`
}

type Data struct {
	X int64     `json:"x"`
	Y []float64 `json:"y"`
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

var offset int

func NewAnnotation(x int64, color, text string) Annotation {
	if offset < 288 {
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

func NewChartData(userID uint, productID string, alpha, omega int64) Chart {

	rates := model.GetAllRatesBetween(userID, productID, alpha, omega)

	pattern := model.GetPattern(userID, productID)

	user := model.FindUserByID(userID)

	var gain float64
	var loss float64
	var tup float64
	var tdn float64

	var annos []Annotation

	var then, that model.Rate
	for i, this := range rates {

		if pattern.MatchesTweezerBottomPattern(then, that, this) {

			var entry, exit, goal, entryPlusFee float64
			var lastRate model.Rate

			for j, rate := range rates[i:] {

				if j == 0 {
					exit = -1.0
					entry = rate.Open
					goal = pattern.GoalPrice(entry)
					entryPlusFee = entry + (entry * user.Taker)
				}

				lastRate = rate

				// if the low meets or exceeds our loss limit ...
				if rate.Low <= pattern.Loss {

					// ok, not the worst thing in the world, maybe a stop order already sold this for us
					if exit == -1.0 {
						// nope, we never established a stop order for this chart, we took a bath
						exit = pattern.Loss
					}
					break
				}

				// else if the high meets or exceeds our gain limit ...
				if rate.High >= goal {

					// is this the first time this has happened?
					if exit == -1.0 {
						// great, we have a stop (limit) entry order placed, continue on.
						exit = goal
					}

					// now if the rate closes less than our exit, the entry order would have been triggered.
					if rate.Close < exit {
						break
					}

					// else we're trending up, ride the wave.
					if rate.Close >= exit {
						exit = rate.Close
					}
				}

				// else the low and highs of this rate do not exceed either respective limit
				// we must now navigate each rate and attempt to sell at profit
				// and avoid the ether
				if j == 0 {
					continue
				}

				if exit == 0 && rate.Time().Sub(this.Time()) > time.Minute*75 && rate.High >= entryPlusFee {
					exit = entryPlusFee
					break
				}
			}

			x := this.Time().UTC().Unix()
			a1 := NewAnnotation(x, Enter, util.FloatToUsd(entry))
			annos = append(annos, a1)
			x2 := lastRate.Time().UTC().Unix()

			if exit > 0 {

				exitPlusFee := exit + (exit * user.Maker)

				result := exitPlusFee - entryPlusFee

				if result >= 0 {
					gain += result
					a2 := NewAnnotation(x2, Won, util.FloatToUsd(exit))
					annos = append(annos, a2)
				} else {
					loss += result
					a2 := NewAnnotation(x2, Lost, util.FloatToUsd(exit))
					annos = append(annos, a2)
				}

			} else {
				a1 := NewAnnotation(x2, Trade, util.FloatToUsd(lastRate.Close))
				annos = append(annos, a1)
				if lastRate.Close > entry {
					tup += entry - lastRate.Close
				} else {
					tdn += entry - lastRate.Close
				}
			}
		}
		then = that
		that = this
	}

	title := fmt.Sprintf("%s=%f ... %s=%f ... %s=%f ... %s=%f",
		util.Won, gain,
		util.Lost, loss,
		util.TradingUp, tup,
		util.TradingDown, tdn,
	)

	var data []Data
	for _, rate := range rates {
		data = append(data, Data{
			X: rate.Time().UTC().Unix(),
			Y: rate.Data(),
		})
	}

	return Chart{
		Series: []Series{{"candles", data}},
		Options: Options{
			Detail:      detail,
			Title:       Title{title, "left"},
			Tooltip:     tooltip,
			XAxis:       xaxis,
			YAxis:       yaxis,
			Annotations: Annotations{annos},
		},
	}
}
