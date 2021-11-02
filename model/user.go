package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"gorm.io/gorm"
	"net/http"
	"nuchal-api/db"
	"time"
)

type User struct {
	gorm.Model
	Api
	Name  string `json:"name"`
	Image string `json:"image"`
}

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

func init() {
	db.Migrate(&User{})
}

func FindUserByID(ID uint) User {

	var user User

	db.Resolve().
		Where("id = ?", ID).
		Find(&user)

	return user
}

func FindUsers() []User {
	var users []User
	db.Resolve().Find(&users)
	return users
}

func SaveUser(user User) User {
	if user.ID > 0 {
		db.Resolve().Save(&user)
	} else {
		db.Resolve().Create(&user)
	}
	return user
}

func DeleteUser(ID uint) {
	db.Resolve().Delete(&User{}, ID)
}
