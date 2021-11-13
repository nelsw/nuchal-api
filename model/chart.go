package model

type Chart struct {
	Layer  Layer   `json:"chart"`
	Layers []Layer `json:"onchart"`
}

type LayerType string

const (
	orderLayer    LayerType = "Trades"
	candleLayer             = "Candles"
	splitterLayer           = "Splitters"
)

type Layer struct {
	Type     LayerType       `json:"type"`
	Name     string          `json:"name"`
	Data     [][]interface{} `json:"data"`
	Settings Settings        `json:"settings"`
}

type Settings struct {
	Legend    bool    `json:"legend"`
	ZIndex    int     `json:"z-index"`
	LineWidth float64 `json:"line_width;omitempty"`
}

func NewProductChart(userID uint, productID string, alpha, omega int64) (chart Chart, err error) {

	var rates []Rate
	if rates, err = GetRates(userID, productID, alpha, omega); err != nil {
		return
	}

	var data [][]interface{}
	for _, rate := range rates {
		data = append(data, rate.data())
	}

	chart.Layer = Layer{candleLayer, "", data, Settings{}}
	chart.Layers = []Layer{
		{orderLayer, "Orders", [][]interface{}{}, Settings{Legend: false, ZIndex: 5}},
		{splitterLayer, "Splits", [][]interface{}{}, Settings{Legend: false, ZIndex: 10}},
	}

	return
}
