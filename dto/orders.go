package dto

import (
	"encoding/json"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"time"
)

type Order struct {
	ID            string            `json:"id"`
	UserID        string            `json:"user_id"`
	Market        string            `json:"market"`
	Side          model.Side        `json:"side"`
	Price         float64           `json:"price"`
	OriginalSize  float64           `json:"original_size"`
	RemainingSize float64           `json:"remaining_size"`
	QuoteAmount   float64           `json:"quote_amount"`
	AvgDealtPrice float64           `json:"avg_dealt_price"`
	Type          book.OrderType    `json:"type"`
	Mode          model.Mode        `json:"mode"`
	Status        model.OrderStatus `json:"status"`
	CreatedAt     time.Time         `json:"-"`
	UpdatedAt     time.Time         `json:"-"`
}

func (o Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		*Alias
		CreatedAt int64 `json:"created_at"`
		UpdatedAt int64 `json:"updated_at"`
	}{
		Alias:     (*Alias)(&o),
		CreatedAt: o.CreatedAt.UnixMilli(),
		UpdatedAt: o.UpdatedAt.UnixMilli(),
	})
}
