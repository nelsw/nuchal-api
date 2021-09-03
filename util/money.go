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

func FloatToDecimal(f float64) string {
	s := fmt.Sprintf("%f", f)
	return StringToDecimal(s)
}

func StringToDecimal(s string) string {

	chunks := strings.Split(s, `.`)

	dollars := chunks[0]

	var cents string
	if len(chunks) > 1 {
		cents = chunks[1]
		cents = strings.TrimRight(cents, "0")
	}

	isNegative := strings.Contains(dollars, "-")
	if isNegative {
		chunks = strings.Split(dollars, "-")
		dollars = chunks[1]
	}

	places := len(dollars)

	pivot := places - 3
	var newFields []string
	for i, oldField := range dollars {
		if pivot != 0 && i != 0 && i%pivot == 0 {
			newFields = append(newFields, ",")
		}
		newFields = append(newFields, string(oldField))
	}

	result := strings.Join(newFields, ``)
	if isNegative {
		result = "-"
		return fmt.Sprintf("-%s.%s", result, cents)
	}

	return fmt.Sprintf("%s.%s", result, cents)
}
