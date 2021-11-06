package model

import (
	"fmt"
	"nuchal-api/util"
	"testing"
	"time"
)

func TestNewProductSim(t *testing.T) {

	fmt.Println(time.Unix(1636082173, 0))
	fmt.Println(time.Unix(1633403773, 0))

	response := NewProductSim(uint(1), uint(1), 1633403773, 1636082173)
	util.PrettyPrint(response)

}
