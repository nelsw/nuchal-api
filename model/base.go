package model

import (
	"gorm.io/gorm"
)

type StrModel struct {
	ID        string         `json:"id" gorm:"primarykey"`
	CreatedAt int64          `json:"created_at" gorm:"autoCreateTime:nano"`
	UpdatedAt int64          `json:"updated_at" gorm:"autoUpdateTime:nano"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"autoDeleteTime:nano;index"`
}

type UintModel struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt int64          `json:"created_at" gorm:"autoCreateTime:nano"`
	UpdatedAt int64          `json:"updated_at" gorm:"autoUpdateTime:nano"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"autoDeleteTime:nano;index"`
}
