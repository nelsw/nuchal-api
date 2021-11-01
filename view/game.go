package view

import (
	"github.com/rs/zerolog/log"
	"nuchal-api/model"
	"sort"
	"strings"
	"sync"
)

const (
	headFmt = "%s\t%s\t%s\t%s\t%s"
	lineFmt = "%s\t%f\t%f\t%f\t%f"
)

type Game struct {
	ProductID    string    `json:"product_id"`
	Currency     string    `json:"currency"`
	Plays        []Play    `json:"-"`
	Size         int       `json:"size"`
	BestRatio    float64   `json:"best_ratio"`
	BestBonus    float64   `json:"best_bonus"`
	AverageRatio float64   `json:"average_ratio"`
	AverageBonus float64   `json:"average_bonus"`
	Payloads     []Payload `json:"payloads"`
}

type Play struct {
	Rates []model.Rate
	Bonus float64
	Ratio float64
}

func (p Play) hasValue() bool {
	return len(p.Rates) > 3 && p.Bonus > 0
}

func FindAllGames() []Game {
	var games []Game
	var wg sync.WaitGroup
	for _, productID := range model.ProductIDs {
		wg.Add(1)
		go func(productID string) {
			defer wg.Done()
			game := FindGame(productID)
			if game != nil {
				games = append(games, *game)
			}
		}(productID)
	}
	wg.Wait()
	return games
}

func FindGame(productID string) *Game {

	plays := FindPlays(productID)
	size := len(plays)
	if size < 1 {
		return nil
	}

	var bestRatio, bestBonus, averageRatio, averageBonus float64
	for _, play := range plays {
		if play.Ratio > bestRatio {
			bestRatio = play.Ratio
		}
		if play.Bonus > bestBonus {
			bestBonus = play.Bonus
		}
		averageRatio += play.Ratio
		averageBonus += play.Bonus
	}

	averageRatio /= float64(size)
	averageBonus /= float64(size)

	var payloads []Payload
	for i, play := range plays {
		var data [][]interface{}
		for _, rate := range play.Rates {
			data = append(data, rate.OHLCV())
		}
		payloads = append(payloads, NewPayload(productID, data))
		if i > 8 {
			break
		}
	}

	return &Game{
		ProductID:    productID,
		Currency:     strings.ReplaceAll(productID, "-USD", ""),
		Plays:        plays,
		Size:         size,
		BestRatio:    bestRatio,
		BestBonus:    bestBonus,
		AverageRatio: averageRatio,
		AverageBonus: averageBonus,
		Payloads:     payloads,
	}
}

func FindPlays(productID string) []Play {

	var plays []Play

	rates := model.FindAllRates(productID)
	sort.SliceStable(rates, func(i, j int) bool {
		return rates[i].Time().Before(rates[j].Time())
	})

	for i, rate := range rates {
		if rate.Close >= rate.Open {
			if play := findPlay(rates, i, rate); play.hasValue() {
				plays = append(plays, play)
			}
		}
	}

	sort.SliceStable(plays, func(i, j int) bool {
		return plays[i].Ratio > plays[j].Ratio
	})

	log.Trace().Str("productID", productID).Int("plays", len(plays)).Msg("FindPlays")

	return plays
}

func findPlay(rates []model.Rate, start int, that model.Rate) Play {

	var min = that.Open
	var end int
	var this model.Rate

	for end, this = range rates[start+1:] {
		if that.Open > this.Open || this.Open > this.Low {
			break
		}
		min = this.Open
		that = this
	}

	return Play{
		rates[start : start+2+end],
		that.Close - min,
		(that.Close - min) / min * 100.0,
	}
}
