package view

import (
	"nuchal-api/model"
)

type Pattern struct {
	model.Pattern
	Product model.Product `json:"product"`
}

func GetPatterns(userID uint) []Pattern {

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
	p := model.GetPattern(userID, productID)
	return Pattern{
		Pattern: p,
		Product: model.ProductMap[p.ProductID],
	}
}
