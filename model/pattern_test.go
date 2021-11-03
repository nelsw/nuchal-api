package model

import (
	"fmt"
	"nuchal-api/util"
	"testing"
)

func TestGetPatterns(t *testing.T) {

	userID := uint(1)

	InitProducts(userID)

	//fmt.Println(len(ProductArr))

	patterns := GetPatterns(userID)
	//for _, product := range ProductArr {
	//	util.PrettyPrint(product)
	//}

	fmt.Println(patterns[0].Wat(2.01))

	fmt.Println(util.Pretty(patterns))

}

//
//func TestSavePattern(t *testing.T) {
//	pattern := GetPattern(uint(1), uint(6))
//	fmt.Println(util.Pretty(pattern))
//	pattern.Enable = true
//	pattern.Save()
//	fmt.Println(util.Pretty(GetPattern(uint(1), uint(6))))
//}
