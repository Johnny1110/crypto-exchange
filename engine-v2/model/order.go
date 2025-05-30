package model

import (
	"time"
)

type Side int
type Mode int
type OrderStatus string

const (
	BID Side = iota
	ASK
)

const (
	MAKER Mode = iota
	TAKER
)

const (
	// ORDER_STATUS_NEW indicates an order that has just been created and not yet matched.
	ORDER_STATUS_NEW OrderStatus = "NEW"

	// ORDER_STATUS_PARTIAL indicates an order that has been partially filled.
	ORDER_STATUS_PARTIAL OrderStatus = "PARTIAL"

	// ORDER_STATUS_FILLED indicates an order that has been completely filled.
	ORDER_STATUS_FILLED OrderStatus = "FILLED"

	// ORDER_STATUS_CANCELED indicates an order that has been canceled.
	ORDER_STATUS_CANCELED OrderStatus = "CANCELED"
)

type Order struct {
	ID                   string
	UserID               string
	Side                 Side
	Price                float64
	OriginalSize         float64
	RemainingSize        float64
	OriginalQuoteAmount  float64
	RemainingQuoteAmount float64
	Mode                 Mode
	Timestamp            time.Time
}

func (o *Order) GetStatus() OrderStatus {
	if o.OriginalSize == o.RemainingSize {
		return ORDER_STATUS_NEW
	}
	if o.RemainingSize > 0 && o.RemainingSize < o.OriginalSize {
		return ORDER_STATUS_PARTIAL
	}
	if o.RemainingSize == 0 {
		return ORDER_STATUS_FILLED
	}
	return ORDER_STATUS_CANCELED
}

// NewOrder
// side: BID ASK
// mode: MAKER TAKER
func NewOrder(orderId, userId string, side Side, price float64, size float64, mode Mode) *Order {
	return &Order{
		ID:            orderId,
		UserID:        userId,
		Side:          side,
		Price:         price,
		OriginalSize:  size,
		RemainingSize: size,
		Mode:          mode,
		Timestamp:     time.Now(),
	}
}

type OrderNode struct {
	Order      *Order
	Prev, Next *OrderNode
}

func NewOrderNode(orderId, userId string, side Side, price float64, qty float64, orderType Mode) *OrderNode {
	order := NewOrder(orderId, userId, side, price, qty, orderType)
	return &OrderNode{
		Order: order,
	}
}

func (node *OrderNode) Size() float64 {
	return node.Order.RemainingSize
}

func (node *OrderNode) Price() float64 {
	return node.Order.Price
}
