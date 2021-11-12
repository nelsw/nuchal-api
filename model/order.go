package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
	"time"
)

type SellOrder struct {
	ID                  string  `json:"id"`
	CreatedAtUnixSecond int64   `json:"created_at"`
	Price               float64 `json:"price"`
	PriceText           string  `json:"price_text"`
	Size                float64 `json:"size"`
	SizeText            string  `json:"size_text"`
}

// CreateOrder creates an order on Coinbase and returns the order once it is no longer pending and has settled.
func CreateOrder(pattern Pattern, order cb.Order, attempt ...int) (cb.Order, error) {

	u := FindUserByID(pattern.UserID)

	pattern.Logger().Info().Str("orderID", order.ID).Msg("create order")

	r, err := u.Client().CreateOrder(&order)
	if err == nil {

		pattern.Logger().Info().Str("orderID", order.ID).Msg("created order")

		return GetOrder(pattern, r.ID)
	}

	if err != nil {
		pattern.Logger().Error().Err(err).Str("orderID", order.ID).Msg("order")
	}

	i := util.FirstIntOrZero(attempt)
	if err.Error() == "Insufficient funds" || i > 10 {
		return cb.Order{}, err
	}

	i++
	time.Sleep(time.Duration(i*3) * time.Second)
	return CreateOrder(pattern, order, i)
}

// GetOrder is a recursive function that returns an order equal to the given id once it is settled and not pending.
func GetOrder(pattern Pattern, orderID string, attempt ...int) (cb.Order, error) {

	u := FindUserByID(pattern.UserID)

	pattern.Logger().Info().Str("orderID", orderID).Msg("get order")

	order, err := u.Client().GetOrder(orderID)

	if err != nil {

		i := util.FirstIntOrZero(attempt)

		pattern.Logger().Error().
			Err(err).
			Str("orderID", orderID).
			Int("attempt", i).
			Msg("error getting order")

		if i > 10 {
			return cb.Order{}, err
		}

		i++
		time.Sleep(time.Duration(i*3) * time.Second)
		return GetOrder(pattern, orderID, i)
	}

	if !order.Settled || order.Status == "pending" {

		pattern.Logger().Warn().
			Str("orderID", orderID).
			Str("side", order.Side).
			Str("type", order.Type).
			Msg("got order, but it's pending or unsettled")

		time.Sleep(1 * time.Second)
		return GetOrder(pattern, orderID, 0)
	}

	pattern.Logger().Info().
		Str("orderID", orderID).
		Str("side", order.Side).
		Str("type", order.Type).
		Msg("got order")

	return order, nil
}

// CancelOrder is a recursive function that cancels an order equal to the given id.
func CancelOrder(pattern Pattern, orderID string, attempt ...int) error {

	u := FindUserByID(pattern.UserID)

	pattern.Logger().Info().Str("orderID", orderID).Msg("cancel order")

	err := u.Client().CancelOrder(orderID)
	if err == nil {
		pattern.Logger().Info().Str("orderID", orderID).Msg("canceled order")
		return nil
	}

	i := util.FirstIntOrZero(attempt)
	pattern.Logger().Error().
		Err(err).
		Str("orderID", orderID).
		Int("attempt", i).
		Msg("error canceling order")

	if i > 10 {
		return err
	}

	i++
	time.Sleep(time.Duration(i*3) * time.Second)

	return CancelOrder(pattern, orderID, i)
}

func GetOrders(userID uint, product Product) ([]SellOrder, error) {
	u := FindUserByID(userID)
	var sellOrders []SellOrder
	var orders []cb.Order
	cursor := u.Client().ListOrders(cb.ListOrdersParams{ProductID: product.ID})
	for cursor.HasMore {
		if err := cursor.NextPage(&orders); err != nil {
			return nil, err
		}
		for _, order := range orders {
			sellOrders = append(sellOrders, SellOrder{
				ID:                  order.ID,
				CreatedAtUnixSecond: order.CreatedAt.Time().UTC().Unix(),
				Price:               util.StringToFloat64(order.Price),
				PriceText:           "$" + product.precise(util.StringToFloat64(order.Price)),
				Size:                util.StringToFloat64(order.Size),
				SizeText:            product.precise(util.StringToFloat64(order.Size)),
			})
		}

	}
	return sellOrders, nil
}

func PostOrder(userID uint, o cb.Order) (err error) {
	log.Trace().Uint("userID", userID).Interface("order", o).Send()
	u := FindUserByID(userID)
	_, err = u.Client().CreateOrder(&o)
	if err != nil {
		log.Error().Err(err).Stack().Uint("userID", userID).Interface("order", o).Send()
	}
	return
}

func DeleteOrder(userID uint, orderID string) (err error) {
	u := FindUserByID(userID)
	err = u.Client().CancelOrder(orderID)
	return
}
