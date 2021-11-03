package model

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Bate struct {
	Base      string `json:"base" gorm:"primaryKey"`
	Quote     string `json:"quote" gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (b Bate) ID() string {
	return fmt.Sprintf("%s-%s", b.Base, b.Quote)
}
