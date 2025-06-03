package dto

import "time"

type PlaceOrderResult struct {
	Matches []*Match `json:"matches"`
	Order   Order    `json:"order"`
}

type Match struct {
	Price     float64   `json:"price"`
	Size      float64   `json:"size"`
	Timestamp time.Time `json:"timestamp"`
}
