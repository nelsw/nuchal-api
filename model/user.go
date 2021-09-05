package model

import (
	"gorm.io/gorm"
	"nuchal-api/db"
)

type User struct {
	gorm.Model
	Api
	Name  string `json:"name"`
	Image string `json:"image"`
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
