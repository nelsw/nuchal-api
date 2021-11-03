package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"nuchal-api/db"
)

type Favorite struct {
	gorm.Model
	UserID      uint `json:"user_id"`
	cb.Currency `gorm:"embedded;embeddedPrefix:currency_"`
	cb.Product  `gorm:"embedded;embeddedPrefix:product_"`
	cb.Ticker   `gorm:"embedded;embeddedPrefix:ticker_"`
	cb.Stats    `gorm:"embedded;embeddedPrefix:stats_"`
}

//func RefreshFavorites(userID uint) ([]Favorite, error) {
//
//}

func GetFavorites(userID uint) []Favorite {
	var favorites []Favorite
	tx := db.Resolve().Find(&favorites)
	log.Trace().Interface("tx", tx)
	return favorites
}

//func NewFavorite(userID uint, baseCurrency string) (Favorite, error) {
//
//
//
//}
