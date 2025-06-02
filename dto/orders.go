package dto

import (
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
)

type Order struct {
	ID            string            `json:"id"`
	UserID        string            `json:"user_id"`
	Market        string            `json:"market"`
	Side          model.Side        `json:"side"`
	Price         float64           `json:"price"`
	OriginalSize  float64           `json:"original_size"`
	RemainingSize float64           `json:"remaining_size"`
	Type          book.OrderType    `json:"type"`
	Mode          model.Mode        `json:"mode"`
	Status        model.OrderStatus `json:"status"`
	CreatedAt     int64             `json:"created_at"`
	UpdatedAt     int64             `json:"updated_at"`
}
