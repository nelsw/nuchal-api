package service

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"nuchal-api/model"
	"nuchal-api/util"
	"sort"
)

func GetRemainingBuyFills(userID uint, productID string) ([]cb.Fill, error) {

	u := model.FindUserByID(userID)

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
		return allFills[i].CreatedAt.Time().Before(allFills[j].CreatedAt.Time())
	})

	var buys, sells []cb.Fill

	for _, fill := range allFills {
		if fill.Side == "buy" {
			buys = append(buys, fill)
		} else {
			sells = append(sells, fill)
		}
	}

	qty := util.MinInt(len(buys), len(sells))
	result := buys[qty:]

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].CreatedAt.Time().After(result[j].CreatedAt.Time())
	})

	return result, nil
}
