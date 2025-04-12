package main

import (
	"fmt"
	"sort"
	"time"
)

type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

// <Order> ----------------------------------------------------------
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

// <Limit> ----------------------------------------------------------
// Limit is a group of orders at the same price level
type Limit struct {
	Price       float64
	Orders      Orders
	TotalVolume float64
}

func (limit *Limit) FillOrder(inputOrder *Order) []Match {
	matches := []Match{}

	for _, existingOrder := range limit.Orders {
		match := limit.fillOrder(existingOrder, inputOrder)
		matches = append(matches, match)
	}

	return matches
}

func (limit *Limit) fillOrder(orderA, orderB *Order) Match {
	// TODO: implement the logic to fill the order (2025/04/12)
	panic("unimplemented")
}

func (limit *Limit) String() string {
	return fmt.Sprintf("Limit:<Price: %.2f, TotalVolume: %.2f>", limit.Price, limit.TotalVolume)
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

	// resort the rest of the orders
	sort.Sort(l.Orders)
}

// <OrderBook> ----------------------------------------------------------
type OrderBook struct {
	asks []*Limit
	bids []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		asks:      []*Limit{},
		bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
	}
}

func (orderBook *OrderBook) PlaceLimitOrder(price float64, order *Order) {
	var limit *Limit

	if order.Bid {
		limit = orderBook.BidLimits[price]
	} else {
		limit = orderBook.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)

		if order.Bid {
			orderBook.bids = append(orderBook.bids, limit)
			orderBook.BidLimits[price] = limit
		} else {
			orderBook.asks = append(orderBook.asks, limit)
			orderBook.AskLimits[price] = limit
		}
	}

	limit.AddOrder(order)
}

func (orderBook *OrderBook) PlaceMarketOrder(order *Order) []Match {
	var matches = make([]Match, 0)

	if order.Bid {
		// buy order need looking for ask limits (Asks() will be ordered by price)
		for _, limit := range orderBook.Asks() {
			matches = limit.FillOrder(order)
		}

	} else {

	}

	return matches
}

func (ob *OrderBook) Asks() []*Limit {
	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}

func (ob *OrderBook) Bids() []*Limit {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
}

// ------------------------------------------------------------------------

// Limits
type Limits []*Limit

// ByBestAsk
type ByBestAsk struct{ Limits }

func (a ByBestAsk) Len() int           { return len(a.Limits) }
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }
func (a ByBestAsk) Swap(i, j int)      { a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }

// ByBestBid
type ByBestBid struct{ Limits }

func (b ByBestBid) Len() int           { return len(b.Limits) }
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price > b.Limits[j].Price }
func (b ByBestBid) Swap(i, j int)      { b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i] }

// Orders
type Orders []*Order

func (orders Orders) Len() int { return len(orders) }
func (orders Orders) Swap(i, j int) {
	(orders)[i], (orders)[j] = (orders)[j], (orders)[i]
}
func (orders Orders) Less(i, j int) bool {
	return orders[i].Timestamp < (orders)[j].Timestamp
}
