package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"time"
)

type Coinbase struct {
	Refreshed  time.Time
	Products   []cb.Product
	Currencies []cb.Currency
}

func InitCoinbase() {

}
