package db

import (
	"fmt"
	"nuchal-api/model"
	"nuchal-api/util"
	"testing"
)

func TestResolve(t *testing.T) {
	var u []model.User
	db := Resolve()
	db.Find(&u)
	fmt.Println(util.Pretty(&u))
}
