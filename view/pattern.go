package view

import (
	"nuchal-api/model"
)

type Pattern struct {
	model.Pattern
	model.Product `json:"product"`
}

func GetPatterns(userID uint) []Pattern {

	_ = model.InitProducts(userID)

	var patterns []Pattern
	for _, p := range model.GetPatterns(userID) {
		pattern := Pattern{
			Pattern: p,
			Product: model.ProductMap[p.ProductID],
		}
		patterns = append(patterns, pattern)
	}

	return patterns
}

func GetPattern(userID uint, productID string) Pattern {
	return Pattern{model.GetPattern(userID, productID), model.ProductMap[productID]}
}
