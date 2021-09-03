package model

type Price struct {
	Unix      int64   `json:"unix" gorm:"primaryKey"`
	ProductID string  `json:"product_id" gorm:"primaryKey"`
	Price     float64 `json:"price"`
	Size      int     `json:"size"`
	Side      string  `json:"side"`
}
