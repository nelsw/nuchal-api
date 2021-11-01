package view

import (
	"nuchal-api/model"
)

func FindRatesBetween(productID string, alpha, omega int64) Response {
	var data [][]interface{}
	for _, rate := range model.FindRatesBetween(productID, alpha, omega) {
		data = append(data, rate.OHLCV())
	}
	return Response{Result{candle, Settings{}, productID, data}, nil, Analysis{}}
}
