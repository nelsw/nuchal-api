package view

import (
	"fmt"
	"nuchal-api/model"
	"nuchal-api/util"
	"strconv"
	"time"
)

type Simulation struct {

	// Gain is the gross amount won, fees not included.
	Gain float64 `json:"gain"`

	// Loss is the gross amount lost, fees not included.
	Loss float64 `json:"loss"`

	// Rise are the trade we're holding which are currently greater than the buy.
	Rise float64 `json:"rise"`

	// Fall are the trade we're holding which are currently less than the buy.
	Fall float64 `json:"fall"`

	// Stake is the amount we invested.
	Stake float64 `json:"stake"`

	// Take is the result of summing the total with closed holds
	Take float64 `json:"take"`

	// Net profit or loss, fees included. Does not include held products.
	Net float64 `json:"net"`

	// Total is the net less maker and taker fees.
	Total float64 `json:"total"`

	// Transactions are the trade that comprise the simulation.
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Index int `json:"index"`

	// Alpha is the time of the buy.
	Alpha int64 `json:"alpha"`

	// Omega is the time of the sell.
	Omega int64 `json:"omega"`

	// Buy is the price which we purchase the product.
	Buy float64 `json:"buy"`

	// Sell is the price which we want to sell the product.
	Sell float64 `json:"sell"`

	// Sold is the price which we sold the product.
	Sold float64 `json:"sold"`

	// Hold is the last known price of the tradeType, it's still active.
	Hold float64 `json:"hold"`

	// Goal is the price which we want to sell that product including fees.
	Goal float64 `json:"goal"`

	// Stop is the price which we sell no matter what.
	Stop float64 `json:"stop"`

	// Entry is the buy price plus maker fees.
	Entry float64 `json:"entry"`

	// Exit is sell price plus taker fees.
	Exit float64 `json:"exit"`

	// Result is the net less maker and taker fees.
	Result float64 `json:"result"`

	AveragePrices []float64 `json:"average_prices"`
}

func NewSimulation(userID uint, productID string, alpha, omega int64) (string, Simulation, []model.Data, []Annotation) {

	rates := model.GetAllRatesBetween(userID, productID, alpha, omega)

	pattern := model.GetPattern(userID, productID)

	user := model.FindUserByID(userID)

	var gain, loss, holdUp, holdDown, stake float64

	var annotations []Annotation
	var transactions []Transaction

	var then, that model.Rate
	var index = 0
	for i, this := range rates {

		if pattern.Bound == "Buys" && len(transactions) == pattern.Break {
			break
		}

		if pattern.MatchesTweezerBottomPattern(then, that, this) {

			var averagePrices []float64

			var buy, sell, sold, goal, entry, stop float64
			var lastRate model.Rate

			for j, rate := range rates[i:] {

				averagePrices = append(averagePrices, rate.AveragePrice())

				if j == 0 {
					sold = -1.0
					buy = rate.Open
					stake += buy * pattern.Size
					sell = pattern.GoalPrice(buy)
					stop = pattern.LossPrice(buy)
					goal = sell + (sell * user.Maker)
					entry = buy + (buy * user.Taker)
				}

				lastRate = rate

				// if the low meets or exceeds our loss limit ...
				if rate.Low < stop {

					// ok, not the worst thing in the world, maybe a stop order already sold this for us
					if sold < 0 {
						// nope, we never established a stop order for this chart, we took a bath
						sold = stop
					}
					break
				}

				// else if the high meets or exceeds our gain limit ...
				if rate.High >= goal {

					// is this the first time this has happened?
					if sold < 0 {
						// great, we have a stop (limit) buy order placed, continue on.
						sold = sell
					}

					// now if the rate closes less than our sold, the buy order would have been triggered.
					if rate.Close < goal {
						break
					}

					// else we're trending up, ride the wave.
					if rate.Close >= goal {
						sold = rate.Close
					}
				}

				// else the low and highs of this rate do not exceed either respective limit
				// we must now navigate each rate and attempt to sell at profit
				// and avoid the ether
				if j == 0 {
					continue
				}

				// if we have yet to sell this product
				if sold < 0 &&
					// and we've been holding for some time
					rate.Time().Sub(this.Time()) > time.Hour*12 &&
					// and we can break even ...
					rate.High >= entry+(entry*user.Maker) {
					// sell it
					sold = entry
					break
				}
			}

			alpha := this.Time().UTC().Unix()
			omega := lastRate.Time().UTC().Unix()

			var icon string
			var hold, exit, result float64

			entry *= pattern.Size

			if sold < 0 {
				icon = T
				hold = lastRate.Close
				exit = (lastRate.Close - (lastRate.Close * user.Maker)) * pattern.Size
				result = exit - entry
				if hold >= buy {
					holdUp += hold
				} else {
					holdDown += hold
				}
			} else {
				exit = (sold - (sold * user.Maker)) * pattern.Size
				result = exit - entry
				if result > -0 {
					gain += result
					icon = Won
				} else {
					loss += result * -1
					icon = Lost
				}
			}

			index += 1

			annotations = append(annotations, NewAnnotation(alpha, Enter, strconv.Itoa(index)))
			annotations = append(annotations, NewAnnotation(omega, icon, strconv.Itoa(index)))
			transactions = append(transactions, Transaction{
				Index:         index,
				Alpha:         alpha,
				Omega:         omega,
				Buy:           buy,
				Sell:          sell,
				Stop:          stop,
				Sold:          sold,
				Hold:          hold,
				Goal:          goal,
				Entry:         entry,
				Exit:          exit,
				Result:        result,
				AveragePrices: averagePrices,
			})
		}
		then = that
		that = this
	}

	title := fmt.Sprintf("%s=%f ... %s=%f ... %s=%f ... %s=%f ... %s=%f",
		util.Steak, stake,
		util.Won, gain,
		util.Poo, loss,
		util.TradingUp, holdUp,
		util.TradingDown, holdDown)

	var data []model.Data
	for _, r := range rates {
		data = append(data, r.Data())
	}

	simulation := Simulation{
		Gain:         gain,
		Loss:         loss,
		Rise:         holdUp,
		Fall:         holdDown,
		Stake:        stake,
		Transactions: transactions,
	}

	return title, simulation, data, annotations
}
