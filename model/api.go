package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"net/http"
	"time"
)

type Api struct {
	Key    string  `json:"key"`
	Pass   string  `json:"pass"`
	Secret string  `json:"secret"`
	Maker  float64 `json:"maker"`
	Taker  float64 `json:"taker"`
}

func (v *Api) Client() *cb.Client {
	return &cb.Client{
		"https://api.pro.coinbase.com",
		v.Secret,
		v.Key,
		v.Pass,
		&http.Client{
			Timeout: 15 * time.Second,
		},
		0,
	}
}
