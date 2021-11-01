package util

import (
	"encoding/json"
	"fmt"
)

func Pretty(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "    ")
	return string(b)
}

func PrettyPrint(v interface{}) {
	fmt.Println(Pretty(v))
}
