package util

import (
	"fmt"
	"strings"
)

func FloatToUsd(f float64) string {
	if f == 0 {
		return "$0.00"
	}
	return "$" + FloatToDecimal(f)
}

func StringToUsd(s string) string {
	return "$" + StringToDecimal(s)
}

func FloatToPrice(f float64) string {
	return StringToPrice(fmt.Sprintf("%f", f))
}

func StringToPrice(s string) string {
	return "$" + StringToDecimal(s)
}

func FloatToDecimal(f float64) string {
	return StringToDecimal(fmt.Sprintf("%f", f))
}

func StringToDecimal(s string) string {

	// split the number by decimal
	parts := strings.Split(s, ".")

	// define parts left of the decimal
	left := parts[0]

	var right string
	if len(parts) > 1 {
		// define parts right of the decimal
		right = parts[1]
	}

	// reverse the numbers left of the decimal
	left = reverseString(left)

	// split digits into an array
	chunks := strings.Split(left, "")

	// get the length of the chunks
	size := len(chunks)

	var arr []string
	for i := size - 1; i >= 0; i-- {
		arr = append(arr, chunks[i])
		if i > 0 && i%3 == 0 {
			arr = append(arr, ",")
		}
	}

	left = strings.Join(arr, "")

	if right == "" {
		return left
	}

	right = strings.TrimRight(right, "0")
	if right == "" {
		return left
	}

	return left + "." + right
}

func reverseString(s string) string {
	runes := []rune(s)
	size := len(runes)
	for i := 0; i < size/2; i++ {
		runes[size-i-1], runes[i] = runes[i], runes[size-i-1]
	}
	return string(runes)
}
