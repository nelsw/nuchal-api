package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
	"sync"
)

type History struct {
	Segments []Segment `json:"segments"`
	Result   float64   `json:"result"`
}

func (h *History) Run(event *zerolog.Event, level zerolog.Level, msg string) {
	event.
		Int("segments", len(h.Segments)).
		Float64("result", h.Result)
}

type Segment struct {
	ProductID string  `json:"product_id"`
	Buys      Audit   `json:"buys"`
	Sells     Audit   `json:"sells"`
	Result    float64 `json:"result"`
	Product   Product `json:"-"`
}

func (s *Segment) Run(event *zerolog.Event, level zerolog.Level, msg string) {
	event.
		Str("productID", s.ProductID).
		Str("result", s.Product.precise(s.Result))
}

type Audit struct {
	Side    string    `json:"side"`
	Fills   []cb.Fill `json:"fills"`
	Gross   float64   `json:"gross"`
	Fees    float64   `json:"fees"`
	Net     float64   `json:"net"`
	Step    float64   `json:"step"`
	Product Product   `json:"-"`
}

func (a *Audit) Run(event *zerolog.Event, level zerolog.Level, msg string) {
	event.
		Str("side", a.Side).
		Int("fills", len(a.Fills)).
		Str("fees", a.Product.precise(a.Fees)).
		Str("gross", a.Product.precise(a.Gross)).
		Str("net", a.Product.precise(a.Net))
}

func GetHistory(userID uint) (*History, error) {

	log.Trace().Msg("get history")

	var history History
	historyLog := log.Hook(&history)
	historyLog.Trace().Msg("history")

	products, err := FindAllProducts()
	if err != nil {
		log.Err(err).Stack().Send()
		return nil, err
	}

	var wg sync.WaitGroup

	for _, product := range products {
		if product.Quote != "USD" {
			continue
		}
		wg.Add(1)
		go getHistory(&wg, userID, product, &history)
		wg.Wait()
		historyLog.Trace().Msg("history")
	}

	return &history, nil
}

func getHistory(wg *sync.WaitGroup, userID uint, product Product, history *History) {
	defer wg.Done()

	fills, err := GetAllFills(userID, product.ID)
	if err != nil {
		log.Err(err).Stack().Send()
		return
	}

	if len(fills) < 1 {
		return
	}

	segment := Segment{
		ProductID: product.ID,
		Product:   product,
		Buys:      Audit{Product: product, Step: product.Step},
		Sells:     Audit{Product: product, Step: product.Step},
	}

	historyLog := log.Hook(history)
	segmentLog := historyLog.Hook(&segment)
	buyLog := segmentLog.Hook(&segment.Buys)
	sellLog := segmentLog.Hook(&segment.Sells)

	segmentLog.Trace().Msg("history -> segment")

	for _, fill := range fills {

		price := util.StringToFloat64(fill.Price)
		size := util.StringToFloat64(fill.Size)
		gross := price * size
		fees := util.StringToFloat64(fill.Fee)
		net := gross - fees

		if fill.Side == "buy" {

			segment.Buys.Side = "buy"
			segment.Buys.Fills = append(segment.Buys.Fills, fill)
			segment.Buys.Fees += fees
			segment.Buys.Gross += gross
			segment.Buys.Net += net
			segment.Result += net

			buyLog.Trace().Msg("history -> segment -> buys ")

		} else {

			segment.Sells.Side = "sell"
			segment.Sells.Fills = append(segment.Sells.Fills, fill)
			segment.Sells.Fees += fees
			segment.Sells.Gross += gross
			segment.Sells.Net += net
			segment.Result -= net

			sellLog.Trace().Msg("history -> segment -> sells")
		}
	}
	segment.Result *= -1
	segmentLog.Trace().Msg("history -> segment")
	history.Segments = append(history.Segments, segment)
	history.Result += segment.Result
}
