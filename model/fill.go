package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"nuchal-api/util"
	"sort"
)

func GetRemainingBuyFills(userID uint, productID string, balance float64) ([]cb.Fill, error) {

	u := FindUserByID(userID)

	cursor := u.Client().ListFills(cb.ListFillsParams{ProductID: productID})

	var newFills, allFills []cb.Fill
	for cursor.HasMore {

		if err := cursor.NextPage(&newFills); err != nil {
			return nil, err
		}

		for _, chunk := range newFills {
			allFills = append(allFills, chunk)
		}
	}

	sort.SliceStable(allFills, func(i, j int) bool {
		return allFills[i].CreatedAt.Time().After(allFills[j].CreatedAt.Time())
	})

	var buys []cb.Fill
	var bal float64
	for _, fill := range allFills {
		if fill.Side == "buyOrder" {
			buys = append(buys, fill)
			bal += util.StringToFloat64(fill.Size)
			if bal == balance {
				break
			}
		}
	}

	return buys, nil
}
