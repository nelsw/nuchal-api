package view

import (
	"github.com/hashicorp/go-uuid"
)

type Wrapper struct {
	Payload Payload `json:"chart"`
}

type Payload struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Name     string          `json:"name"`
	Data     [][]interface{} `json:"data"`
	Settings []interface{}   `json:"settings"`
}

func NewPayload(productID string, data [][]interface{}) Payload {
	id, _ := uuid.GenerateUUID()
	return Payload{id, "Candles", productID, data, nil}
}
