package dto

import "time"

type PlaceOrderResult struct {
	Matches  []*Match `json:"matches"`
	AvgPrice float64  `json:"avg_price"`
	OrderID  string   `json:"order_id"`
}

type Match struct {
	Price     float64   `json:"price"`
	Size      float64   `json:"size"`
	Timestamp time.Time `json:"timestamp"`
}
