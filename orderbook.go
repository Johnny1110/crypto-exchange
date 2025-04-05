package main

import (
	"fmt"
	"time"
)

type Order struct {
	Size      float64
	Bid       bool
	Limit     *Limit // to track t he limit
	Timestamp int64  // unix nano seconds
}

func (o *Order) String() string {
	return "Order{" +
		"Size: " + fmt.Sprintf("%f", o.Size) +
		", Bid: " + fmt.Sprintf("%t", o.Bid) +
		"}"
}

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

// Limit is a group of orders at the same price level
type Limit struct {
	Price       float64
	Orders      []*Order
	TotalVolume float64
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:       price,
		Orders:      []*Order{},
		TotalVolume: 0,
	}
}

func (l *Limit) AddOrder(o *Order) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}

func (l *Limit) DeleteOrder(o *Order) {
	for i := 0; i < len(l.Orders); i++ {
		if l.Orders[i] == o {
			// remove the order from the slice
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
		}
	}
	// gc
	o.Limit = nil
	l.TotalVolume -= o.Size

	// TODO: resort the rest of the orders
}

type OrderBook struct {
	Asks []*Limit
	Bids []*Limit
}
