package model

import (
	"fmt"
	"nuchal-api/util"
	"testing"
)

func TestFindUserByID(t *testing.T) {
	FindUserByID(1)
}

func TestSaveUser(t *testing.T) {

	u := User{
		Api:   Api{},
		Name:  "Some Name",
		Image: "",
	}

	savedUser := SaveUser(u)

	fmt.Println(util.Pretty(savedUser))

}
