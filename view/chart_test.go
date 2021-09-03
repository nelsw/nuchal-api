package view

import (
	"nuchal-api/model"
	"testing"
)

func TestNewChartData(t *testing.T) {

	userID := uint(1)
	model.InitProducts(userID)
	productID := "ALGO-USD"
	alpha := int64(1630394175)
	omega := int64(1630462400)
	NewChartData(userID, productID, alpha, omega)

}
