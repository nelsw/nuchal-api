package model

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/db"
	"nuchal-api/util"
	"strings"
)

type Product struct {
	StrModel
	Name  string  `json:"name"`
	Base  string  `json:"base"`
	Quote string  `json:"quote"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Step  float64 `json:"step"`
	Price float64 `json:"price" gorm:"-"`
}

func init() {
	db.Migrate(&Product{})
	fmt.Println("init")
}

func (p *Product) initPrices(wsConn *ws.Conn) error {

	if price, err := getPrice(wsConn, p.ID); err != nil {
		log.Err(err).
			Stack().
			Str("p.ID", p.ID).
			Msg("error getting price")
		return err
	} else {
		p.Price = price
		return nil
	}
}

func (p *Product) initPrice() error {

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Error().
			Err(err).
			Stack().
			Msg("error while opening websocket connection")
		return err
	}

	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			log.Error().
				Err(err).
				Stack().
				Msg("error closing websocket connection")
		}
	}(wsConn)

	if err := wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", []string{p.ID}}},
	}); err != nil {
		log.Error().
			Err(err).
			Msg("error writing message to websocket")
		return err
	}

	if price, err := getPrice(wsConn, p.ID); err != nil {
		log.Err(err).
			Stack().
			Str("p.ID", p.ID).
			Msg("error getting price")
		return err
	} else {
		p.Price = price
		return nil
	}
}

func FindProductByID(ID string) (Product, error) {

	var product Product

	db.Resolve().
		Where("id = ?", ID).
		Find(&product)

	err := product.initPrice()
	if err != nil {
		log.Err(err).Stack().Send()
	}

	return product, err
}

func FindAllProducts() ([]Product, error) {

	var products []Product

	db.Resolve().
		Order("id asc").
		Find(&products)

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Error().
			Err(err).
			Stack().
			Msg("error while opening websocket connection")
		return nil, err
	}

	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			log.Error().
				Err(err).
				Stack().
				Msg("error closing websocket connection")
		}
	}(wsConn)

	var productIDs []string
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}

	if err := wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", productIDs}},
	}); err != nil {
		log.Error().
			Err(err).
			Msg("error writing message to websocket")
		return nil, err
	}

	var newProducts []Product
	for _, product := range products {
		price, err := getPrice(wsConn, product.ID)
		if err != nil {
			log.Err(err).
				Stack().
				Str("product.ID", product.ID).
				Msg("error getting price")
			continue
		}
		product.Price = price
		newProducts = append(newProducts, product)
	}

	return newProducts, nil
}

func FindAllProductsByQuote(quote string) ([]Product, error) {

	var products []Product

	db.Resolve().
		Where("quote = ?", quote).
		Order("id asc").
		Find(&products)

	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.pro.coinbase.com", nil)
	if err != nil {
		log.Error().
			Err(err).
			Stack().
			Msg("error while opening websocket connection")
		return nil, err
	}

	defer func(wsConn *ws.Conn) {
		if err := wsConn.Close(); err != nil {
			log.Error().
				Err(err).
				Stack().
				Msg("error closing websocket connection")
		}
	}(wsConn)

	var productIDs []string
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}

	if err := wsConn.WriteJSON(&cb.Message{
		Type:     "subscribe",
		Channels: []cb.MessageChannel{{"ticker", productIDs}},
	}); err != nil {
		log.Error().
			Err(err).
			Msg("error writing message to websocket")
		return nil, err
	}

	var newProducts []Product
	for _, product := range products {
		price, err := getPrice(wsConn, product.ID)
		if err != nil {
			log.Err(err).
				Stack().
				Str("product.ID", product.ID).
				Msg("error getting price")
			continue
		}
		product.Price = price
		newProducts = append(newProducts, product)
	}

	return newProducts, nil
}

func (p Product) precise(f float64) string {
	decimal := util.FloatToDecimal(p.Step)
	zeros := len(strings.Split(decimal, ".")[1])
	zeroFormat := fmt.Sprintf(".%df", zeros)
	preciseFormat := "%" + zeroFormat
	return fmt.Sprintf(preciseFormat, f)
}
