package model

import (
	"fmt"
	"testing"
)

func TestGetHistory(t *testing.T) {

	h, err := GetHistory(userID)
	if err != nil {
		t.Fail()
	}

	fmt.Println(h.Result)
}
