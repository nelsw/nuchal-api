package view

import (
	"fmt"
	"nuchal-api/util"
	"testing"
)

func TestGetPatterns(t *testing.T) {

	userID := uint(1)

	patterns := GetPatterns(userID)

	fmt.Println(util.Pretty(patterns))

}
