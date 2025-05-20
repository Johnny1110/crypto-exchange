package model

import (
	"time"
)

type Side int

const (
	BID Side = iota
	ASK Side = iota
)

type Order struct {
	ID           string
	UserID       string
	Side         Side
	Price        float64
	OriginalQty  float64
	RemainingQty float64
	Timestamp    time.Time
}

func NewOrder(orderId, userID string, side Side, price float64, qty float64) *Order {
	return &Order{
		ID:           orderId,
		UserID:       userID,
		Side:         side,
		Price:        price,
		OriginalQty:  qty,
		RemainingQty: qty,
		Timestamp:    time.Now(),
	}
}

type OrderNode struct {
	Order      *Order
	Prev, Next *OrderNode
}

func NewOrderNode(orderId, userID string, side Side, price float64, qty float64) *OrderNode {
	order := NewOrder(orderId, userID, side, price, qty)
	return &OrderNode{
		Order: order,
	}
}

func (node *OrderNode) Qty() float64 {
	return node.Order.RemainingQty
}

func (node *OrderNode) Price() float64 {
	return node.Order.Price
}
