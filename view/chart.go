package view

import (
	"github.com/hashicorp/go-uuid"
	"nuchal-api/model"
)

const (
	Enter = `#651FFF`
	Won   = `#69F0AE`
	Lost  = `#F50057`
	T     = `#4E342E`
)

var (
	detail  = Detail{"candlestick", 500}
	tooltip = Tooltip{true, "dark"}
	xaxis   = XAxis{tooltip, "datetime"}
	yaxis   = YAxis{tooltip}
)

type Chart struct {
	ID         string   `json:"id"`
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
func NewChartData(title string, simulation Simulation, data []model.Data, annotations []Annotation) Chart {
	id, _ := uuid.GenerateUUID()
	return Chart{
		ID:         id,
		Simulation: simulation,
		Series:     []Series{{"candle", data}},
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
