package model

import (
	cb "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
	"time"
)

type ord cb.Order

func (o ord) event() *zerolog.Event {
	return log.Log().Str("orderID", o.ID)
}

func (o ord) l() *zerolog.Logger {
	logger := log.
		With().
		Str("orderID", o.ID).
		Logger()
	return &logger
}

type SeverityHook struct{}

func (h SeverityHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level != zerolog.NoLevel {
		e.Str("severity", level.String())
	}
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

func GetOrders(userID uint, productID string) ([]cb.Order, error) {
	u := FindUserByID(userID)
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
