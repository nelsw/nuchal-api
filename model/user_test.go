package model

import (
	"nuchal-api/util"
	"testing"
)

func TestFindUserByID(t *testing.T) {
	u := FindUserByID(1)
	util.PrettyPrint(u)
}

//func TestSaveUser(t *testing.T) {
//
//	u := User{
//		Api:   Api{},
//		Name:  "Some Name",
//		Image: "",
//	}
//
//	savedUser := SaveUser(u)
//
//	fmt.Println(util.Pretty(savedUser))
//
//}
