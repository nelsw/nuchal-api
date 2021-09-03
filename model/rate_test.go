package model

import (
	"fmt"
	"testing"
	"time"
)

func TestGetAllRatesBetween(t *testing.T) {
	fmt.Println(time.Unix(0, 1630472400))
	fmt.Println(time.Unix(1630462400, 0))
	//rates := GetAllRatesBetween(1, "1INCH-USD", 1630386000000, 1630462400000)
	//fmt.Println(util.Pretty(rates))

}
