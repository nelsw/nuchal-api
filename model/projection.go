package model

type Projection struct {
	Buy         float64 `json:"buy"`
	Sell        float64 `json:"sell"`
	Fees        float64 `json:"fees"`
	ROI         float64 `json:"roi"`
	Percent     float64 `json:"percent"`
	BuyText     string  `json:"buy_text"`
	SellText    string  `json:"sell_text"`
	FeesText    string  `json:"fees_text"`
	ROIText     string  `json:"roi_text"`
	PercentText string  `json:"percent_text"`
	Symbol      string  `json:"symbol"`
}

type Watter func(f float64) string

func (p *Projection) setValues(fun func(f float64) string) {

	p.ROI = p.Sell - p.Buy - p.Fees
	p.Percent = p.ROI / (p.Buy + p.Fees) * 100

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
