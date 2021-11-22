package model

import (
	"fmt"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"gorm.io/gorm"
	"math"
	"nuchal-api/db"
	"nuchal-api/util"
	"strconv"
	"strings"
	"time"
)

type Product struct {
	StrModel
	Name    string  `json:"name"`
	Base    string  `json:"base"`
	Quote   string  `json:"quote"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Step    float64 `json:"step"`
	Fixed   int     `json:"fixed"`
	Posture Posture `json:"posture,omitempty" gorm:"-"`
}

type Posture struct {
	Symbol           string    `json:"symbol"`
	Prices           []float64 `json:"prices"`
	Price            float64   `json:"price"`
	PriceText        string    `json:"price_text"`
	Price24h         float64   `json:"price_24h"`
	Price24hText     string    `json:"price_24h_text"`
	Price24hLow      float64   `json:"price_24h_low"`
	Price24hLowText  string    `json:"price_24h_low_text"`
	Price24hHigh     float64   `json:"price_24h_high"`
	Price24hHighText string    `json:"price_24h_high_text"`
	Change           float64   `json:"change"`
	ChangeText       string    `json:"change_text"`
	Percent          float64   `json:"percent"`
	PercentText      string    `json:"percent_text"`
	Volume24h        float64   `json:"volume_24h"`
	Volume24hText    string    `json:"volume_24h_text"`
}

func init() {
	db.Migrate(&Product{})
}

func (p *Product) AfterFind(tx *gorm.DB) (err error) {

	zeros := 1
	sides := strings.Split(util.FloatToDecimal(p.Step), ".")
	if len(sides) > 1 {
		zeros = len(sides[1])
	}
	p.Fixed = zeros

	var omegaRate Rate

	tx.Where("product_id = ?", p.ID).
		Order("unix_second desc").
		First(&omegaRate)

	omegaTime := omegaRate.Time().UTC()
	alphaTime := omegaTime.Add(time.Hour * -24).UTC()

	var rates []Rate
	tx.Where("product_id = ?", p.ID).
		Where("unix_second between ? and ?", alphaTime.Unix(), omegaTime.Unix()).
		Find(&rates)

	if len(rates) < 1 {
		p.Posture = Posture{
			Price:     omegaRate.Close,
			PriceText: "$" + p.precise(omegaRate.Close),
		}
		return nil
	}

	alphaRate := rates[0]

	min := math.Min(omegaRate.Close, alphaRate.Close)
	max := math.Max(omegaRate.Close, alphaRate.Close)

	var percent float64
	if percent = (max - min) / omegaRate.Close * 100; math.IsNaN(percent) {
		percent = 0
	}

	var change float64
	if change = max - min; math.IsNaN(change) {
		change = 0
	}

	symbol := "-"
	if omegaRate.Close > alphaRate.Close {
		symbol = "+"
	}

	var prices []float64
	var high, low, volume float64
	for _, rate := range rates {
		volume += rate.Volume
		if rate.High > high || high == 0 {
			high = rate.High
		}
		if rate.Low < low || low == 0 {
			low = rate.Low
		}
		prices = append(prices, rate.avg())
	}

	p.Posture = Posture{
		Price:            omegaRate.Close,
		PriceText:        "$" + p.precise(omegaRate.Close),
		Price24hHigh:     high,
		Price24hHighText: "$" + p.precise(high),
		Price24hLow:      low,
		Price24hLowText:  "$" + p.precise(low),
		Price24h:         alphaRate.Close,
		Price24hText:     "$" + p.precise(alphaRate.Close),
		Change:           change,
		ChangeText:       "$" + p.precise(change),
		Percent:          percent,
		PercentText:      fmt.Sprintf("%.2f", percent) + "%",
		Volume24h:        volume,
		Volume24hText:    strconv.Itoa(int(volume)),
		Prices:           prices,
		Symbol:           symbol,
	}

	return nil
}

func (p *Product) NewMarketEntryOrder(size string) cb.Order {
	return cb.Order{
		ProductID: p.ID,
		Side:      "buy",
		Size:      size,
		Type:      "market",
	}
}

func (p *Product) NewMarketExitOrder(size string) cb.Order {
	return cb.Order{
		ProductID: p.ID,
		Side:      "sell",
		Size:      size,
		Type:      "market",
	}
}

func (p *Product) NewStopEntryOrder(size string, price float64) cb.Order {
	return cb.Order{
		ProductID: p.ID,
		Price:     p.precise(price),
		Side:      "sell",
		Size:      size,
		Type:      "limit",
		StopPrice: p.precise(price),
		Stop:      "entry",
	}
}

func (p *Product) StopLossOrder(size string, price float64) cb.Order {
	return cb.Order{
		ProductID: p.ID,
		Price:     p.precise(price),
		Side:      "sell",
		Size:      size,
		Type:      "limit",
		StopPrice: p.precise(price),
		Stop:      "loss",
	}
}

func FindProductByID(ID string) (Product, error) {

	var product Product

	db.Resolve().
		Where("id = ?", ID).
		Find(&product)

	return product, nil
}

func FindAllProducts() ([]Product, error) {

	var products []Product

	db.Resolve().
		Order("id asc").
		Find(&products)

	return products, nil
}

func FindAllProductsByQuote(quote string) ([]Product, error) {

	var products []Product

	db.Resolve().
		Where("quote = ?", quote).
		Order("id asc").
		Find(&products)

	return products, nil
}

func (p *Product) precise(f float64) string {

	decimal := util.FloatToDecimal(p.Step)
	sides := strings.Split(decimal, ".")

	if len(sides) > 1 {
		zeros := len(sides[1])
		zeroFormat := fmt.Sprintf(".%df", zeros)
		preciseFormat := "%" + zeroFormat
		result := fmt.Sprintf(preciseFormat, f)
		if result == "NaN" {
			return "0.0"
		}
		return result
	}

	return sides[0]
}
