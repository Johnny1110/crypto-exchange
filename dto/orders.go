package dto

import (
	"encoding/json"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"time"
)

type Order struct {
	ID            string            `json:"id"`
	UserID        string            `json:"-"`
	Market        string            `json:"market"`
	Side          model.Side        `json:"side"`
	Price         float64           `json:"-"`
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

	// Create the struct with conditional price field
	result := struct {
		*Alias
		Price     *float64 `json:"price,omitempty"`
		CreatedAt int64    `json:"created_at"`
		UpdatedAt int64    `json:"updated_at"`
	}{
		Alias:     (*Alias)(&o),
		CreatedAt: o.CreatedAt.UnixMilli(),
		UpdatedAt: o.UpdatedAt.UnixMilli(),
	}

	// Only include price if it's > 0
	if o.Price > 0 {
		result.Price = &o.Price
	}

	return json.Marshal(result)
}
