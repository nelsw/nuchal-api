package service

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog/log"
	"nuchal-api/model"
	"nuchal-api/util"
	"time"
)

// CreateOrder creates an order on Coinbase and returns the order once it is no longer pending and has settled.
func CreateOrder(userID uint, order cb.Order, attempt ...int) (cb.Order, error) {

	u := model.FindUserByID(userID)

	log.Info().
		Uint("userID", userID).
		Str("orderID", order.ID).
		Msg("create order")

	r, err := u.Client().CreateOrder(&order)
	if err == nil {

		log.Info().
			Uint("userID", userID).
			Str("orderID", order.ID).
			Msg("created order")

		return GetOrder(userID, r.ID)
	}

	if err != nil {
		log.Err(err).
			Uint("userID", userID).
			Str("orderID", order.ID).
			Msg("error creating order")
	}

	i := util.FirstIntOrZero(attempt)
	if err.Error() == "Insufficient funds" || i > 10 {
		return cb.Order{}, err
	}

	i++
	time.Sleep(time.Duration(i*3) * time.Second)
	return CreateOrder(userID, order, i)
}

// GetOrder is a recursive function that returns an order equal to the given id once it is settled and not pending.
func GetOrder(userID uint, orderID string, attempt ...int) (cb.Order, error) {

	u := model.FindUserByID(userID)

	log.Info().
		Uint("userID", userID).
		Str("orderID", orderID).
		Msg("get order")

	order, err := u.Client().GetOrder(orderID)

	if err != nil {

		i := util.FirstIntOrZero(attempt)

		log.Error().
			Err(err).
			Uint("userID", userID).
			Str("orderID", orderID).
			Int("attempt", i).
			Msg("error getting order")

		if i > 10 {
			return cb.Order{}, err
		}

		i++
		time.Sleep(time.Duration(i) * time.Second)
		return GetOrder(userID, orderID, i)
	}

	if !order.Settled || order.Status == "pending" {

		log.Warn().
			Uint("userID", userID).
			Str("product", order.ProductID).
			Str("orderID", orderID).
			Str("side", order.Side).
			Str("type", order.Type).
			Msg("got order, but it's pending or unsettled")

		time.Sleep(1 * time.Second)
		return GetOrder(userID, orderID, 0)
	}

	log.Info().
		Uint("userID", userID).
		Str("product", order.ProductID).
		Str("orderID", orderID).
		Str("side", order.Side).
		Str("type", order.Type).
		Msg("got order")

	return order, nil
}

// CancelOrder is a recursive function that cancels an order equal to the given id.
func CancelOrder(userID uint, orderID string, attempt ...int) error {

	u := model.FindUserByID(userID)

	log.Info().
		Uint("userID", userID).
		Str("orderID", orderID).
		Msg("cancel order")

	err := u.Client().CancelOrder(orderID)
	if err == nil {
		log.Info().
			Uint("userID", userID).
			Str("orderID", orderID).
			Msg("canceled order")
		return nil
	}

	i := util.FirstIntOrZero(attempt)
	log.Error().
		Err(err).
		Uint("userID", userID).
		Str("orderID", orderID).
		Int("attempt", i).
		Msg("error canceling order")

	if i > 10 {
		return err
	}

	i++
	time.Sleep(time.Duration(i) * time.Second)

	return CancelOrder(userID, orderID, i)
}

func GetOrders(userID uint, productID string) ([]cb.Order, error) {
	u := model.FindUserByID(userID)
	var orders, nextOrders []cb.Order
	cursor := u.Client().ListOrders(cb.ListOrdersParams{ProductID: productID})
	for cursor.HasMore {
		if err := cursor.NextPage(&nextOrders); err != nil {
			return nil, err
		}
		orders = append(orders, nextOrders...)
	}
	return orders, nil
}
