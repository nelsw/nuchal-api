package model

import "math"

type Projection struct {

	// Buy is the average buy price for this position.
	Buy float64 `json:"buy"`

	// Even is the price where we break even.
	Even     float64 `json:"even"`
	EvenText string  `json:"even_text"`

	// Sell is the average price we look to sell at.
	Sell float64 `json:"sell"`

	Fees float64 `json:"fees"`

	ROI     float64 `json:"roi"`
	Percent float64 `json:"percent"`

	Place float64 `json:"place"`

	BuyText     string `json:"buy_text"`
	SellText    string `json:"sell_text"`
	FeesText    string `json:"fees_text"`
	ROIText     string `json:"roi_text"`
	PercentText string `json:"percent_text"`
	Symbol      string `json:"symbol"`
}

func (p *Projection) setValues(fun func(f float64) string) {

	p.ROI = p.Sell - p.Buy - p.Fees
	if math.IsNaN(p.ROI) {
		p.ROI = 0
	}

	p.Percent = p.ROI / (p.Buy + p.Fees) * 100
	if math.IsNaN(p.Percent) {
		p.Percent = 0
	}

	p.BuyText = "$" + fun(p.Buy)
	p.SellText = "$" + fun(p.Sell)
	p.FeesText = "$" + fun(p.Fees)
	p.ROIText = "$" + fun(p.ROI)

	if p.ROI >= 0 {
		p.Symbol = "+"
	} else {
		p.Symbol = ""
	}

	p.PercentText = fun(p.Percent) + "%"
}
