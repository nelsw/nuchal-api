package util

import (
	"github.com/rs/zerolog/log"
	"strconv"
)

func StringToFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Err(err).Send()
		return -400
	}
	return f
}

func StringToInt64(s string) int64 {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Err(err).Send()
		return -400
	}
	i64 := int64(i)
	return i64
}

func MinInt(a, z int) int {
	if a < z {
		return a
	} else {
		return z
	}
}

func FirstIntOrZero(arr []int) int {
	if arr != nil && len(arr) > 0 {
		return arr[0]
	}
	return 0
}

func StringToUint(s string) uint {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Err(err).Stack().Send()
		return 0
	}
	return uint(i)
}
