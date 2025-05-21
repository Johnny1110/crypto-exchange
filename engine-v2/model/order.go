package model

import (
	"time"
)

type Side int
type Type int

const (
	BID Side = iota
	ASK
)

const (
	MAKER Type = iota
	TAKER
)

type Order struct {
	ID            string
	UserID        string
	Side          Side
	Price         float64
	OriginalSize  float64
	RemainingSize float64
	Type          Type
	Timestamp     time.Time
}

func NewOrder(orderId, userId string, side Side, price float64, size float64, orderType Type) *Order {
	return &Order{
		ID:            orderId,
		UserID:        userId,
		Side:          side,
		Price:         price,
		OriginalSize:  size,
		RemainingSize: size,
		Type:          orderType,
		Timestamp:     time.Now(),
	}
}

type OrderNode struct {
	Order      *Order
	Prev, Next *OrderNode
}

func NewOrderNode(orderId, userId string, side Side, price float64, qty float64, orderType Type) *OrderNode {
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
