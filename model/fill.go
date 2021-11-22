package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
	"sort"
)

type BuyFill struct {
	ID                  int     `json:"id"`
	CreatedAtUnixSecond int64   `json:"created_at"`
	Size                float64 `json:"size"`
	SizeText            string  `json:"size_text"`
	Price               float64 `json:"price"`
	PriceText           string  `json:"price_text"`
	Fee                 float64 `json:"fee"`
	FeeText             string  `json:"fee_text"`
	ProductID           string  `json:"product_id"`
}

func newBuyFill(i int, fill cb.Fill) BuyFill {
	return BuyFill{
		ID:                  i,
		CreatedAtUnixSecond: fill.CreatedAt.Time().UTC().Unix(),
		Size:                util.StringToFloat64(fill.Size),
		SizeText:            fill.Size,
		Price:               util.StringToFloat64(fill.Price),
		PriceText:           fill.Price,
		Fee:                 util.StringToFloat64(fill.Fee),
		FeeText:             fill.Fee,
		ProductID:           fill.ProductID,
	}
}

func GetAllFills(userID uint, productID string) ([]cb.Fill, error) {

	u := FindUserByID(userID)

	cursor := u.Client().ListFills(cb.ListFillsParams{ProductID: productID})

	var newFills, allFills []cb.Fill
	for cursor.HasMore {

		if err := cursor.NextPage(&newFills); err != nil {
			log.Err(err).Stack().Send()
			return nil, err
		}

		for _, chunk := range newFills {
			allFills = append(allFills, chunk)
		}
	}

	sort.SliceStable(allFills, func(i, j int) bool {
		return allFills[i].CreatedAt.Time().After(allFills[j].CreatedAt.Time())
	})

	return allFills, nil
}

func GetRemainingBuyFills(userID uint, productID string, balance float64) ([]BuyFill, error) {

	allFills, err := GetAllFills(userID, productID)
	if err != nil {
		return nil, err
	}

	var buys []BuyFill
	var bal float64
	for i, fill := range allFills {
		if fill.Side == "buy" {
			buys = append(buys, newBuyFill(i, fill))
			bal += util.StringToFloat64(fill.Size)
			if bal == balance {
				break
			}
		}
	}

	return buys, nil
}
