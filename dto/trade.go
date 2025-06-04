package dto

import (
	"encoding/json"
	"time"
)

type PlaceOrderResult struct {
	Matches []*Match `json:"matches"`
	Order   Order    `json:"order"`
}

type Match struct {
	Price     float64   `json:"price"`
	Size      float64   `json:"size"`
	Timestamp time.Time `json:"-"`
}

func (m Match) MarshalJSON() ([]byte, error) {
	type Alias Match
	return json.Marshal(&struct {
		*Alias
		Timestamp int64 `json:"timestamp"`
	}{
		Alias:     (*Alias)(&m),
		Timestamp: m.Timestamp.UnixMilli(),
	})
}
