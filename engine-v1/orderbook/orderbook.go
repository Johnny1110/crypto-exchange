package orderbook

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type Match struct {
	Ask        *Order  `json:"ask"`
	Bid        *Order  `json:"bid"`
	SizeFilled float64 `json:"size_filled"`
	Price      float64 `json:"price"`
}

// <Order> ----------------------------------------------------------
type Order struct {
	ID        int64
	UserID    int64
	Size      float64
	Bid       bool
	Limit     *Limit // to track the limit
	Timestamp int64  // unix nano seconds
}

func (o *Order) IsFilled() bool {
	return o.Size == 0.0
}

func (o *Order) String() string {
	return "Order{" +
		"Size: " + fmt.Sprintf("%f", o.Size) +
		", Bid: " + fmt.Sprintf("%t", o.Bid) +
		"}"
}

func NewOrder(bid bool, size float64, userId int64) *Order {
	return &Order{
		UserID:    userId,
		ID:        int64(rand.Intn(100000000)),
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

func (limit *Limit) FillOrder(inputMarketOrder *Order) []Match {
	var (
		matches        []Match
		ordersToDelete []*Order
	)

	for _, existingLimitOrder := range limit.Orders {
		match := limit.fillOrder(existingLimitOrder, inputMarketOrder)
		matches = append(matches, match)

		limit.TotalVolume -= match.SizeFilled

		if existingLimitOrder.IsFilled() {
			ordersToDelete = append(ordersToDelete, existingLimitOrder)
		}

		if inputMarketOrder.IsFilled() {
			break
		}
	}

	// delete the filled limit orders
	for _, filledLimitOrder := range ordersToDelete {
		fmt.Println("Deleting order:", filledLimitOrder, " Bid:", filledLimitOrder.Bid, " Size:", filledLimitOrder.Size, " Price:", filledLimitOrder.Limit.Price)
		limit.deleteOrder(filledLimitOrder)
	}

	return matches
}

func (limit *Limit) fillOrder(orderA, orderB *Order) Match {
	var (
		ask        *Order
		bid        *Order
		sizeFilled float64
	)

	if orderA.Bid {
		bid = orderA
		ask = orderB
	} else {
		bid = orderB
		ask = orderA
	}

	if bid.Size >= ask.Size {
		bid.Size -= ask.Size
		sizeFilled = ask.Size
		ask.Size = 0.0
	} else {
		ask.Size -= bid.Size
		sizeFilled = bid.Size
		bid.Size = 0.0
	}

	return Match{
		Bid:        bid,
		Ask:        ask,
		SizeFilled: sizeFilled,
		Price:      limit.Price,
	}
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

func (l *Limit) deleteOrder(o *Order) {
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
	Symbol string
	asks   []*Limit
	bids   []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit

	limitOrderIdMap map[int64]*Order
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol:          symbol,
		asks:            []*Limit{},
		bids:            []*Limit{},
		AskLimits:       make(map[float64]*Limit),
		BidLimits:       make(map[float64]*Limit),
		limitOrderIdMap: make(map[int64]*Order),
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

	orderBook.limitOrderIdMap[order.ID] = order
	limit.AddOrder(order)
}

func (orderBook *OrderBook) PlaceMarketOrder(order *Order) []Match {
	var matches = make([]Match, 0)
	var limitToDelete = make([]*Limit, 0)

	if order.Bid {
		// buy order need looking for ask limits (Asks() will be ordered by price)
		if order.Size > orderBook.AskTotalVolume() {
			panic(fmt.Errorf("Not enough ask vloume [size: %.2f] to fill market order [size: %.2f]", orderBook.AskTotalVolume(), order.Size))
		}

		for _, limit := range orderBook.Asks() {
			match := limit.FillOrder(order)

			for _, m := range match {
				// need remove ask
				if m.Ask.IsFilled() {
					delete(orderBook.limitOrderIdMap, m.Ask.ID)
				}
			}

			matches = append(matches, match...)

			if len(limit.Orders) == 0 {
				limitToDelete = append(limitToDelete, limit)
			}

			if order.IsFilled() {
				break
			}
		}

		for _, limit := range limitToDelete {
			orderBook.clearLimit(false, limit)
		}

	} else {
		// sell order need looking for bid limits (Bids() will be ordered by price)
		if (order.Size) > orderBook.BidTotalVolume() {
			panic(fmt.Errorf("Not enough bid vloume [size: %.2f] to fill market order [size: %.2f]", orderBook.BidTotalVolume(), order.Size))
		}
		for _, limit := range orderBook.Bids() {

			match := limit.FillOrder(order)

			for _, m := range match {
				// need remove bid
				if m.Bid.IsFilled() {
					delete(orderBook.limitOrderIdMap, m.Bid.ID)
				}
			}

			matches = append(matches, match...)

			if len(limit.Orders) == 0 {
				limitToDelete = append(limitToDelete, limit)
			}
			if order.IsFilled() {
				break
			}
		}

		for _, limit := range limitToDelete {
			orderBook.clearLimit(true, limit)
		}
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

func (ob *OrderBook) BidTotalVolume() float64 {
	totalVolumne := 0.0
	for _, limit := range ob.bids {
		totalVolumne += limit.TotalVolume
	}
	return totalVolumne
}

func (ob *OrderBook) AskTotalVolume() float64 {
	totalVolumne := 0.0
	for _, limit := range ob.asks {
		totalVolumne += limit.TotalVolume
	}
	return totalVolumne
}

func (ob *OrderBook) clearLimit(bid bool, limit *Limit) {
	if bid {
		delete(ob.BidLimits, limit.Price)
		// remove limit from orderbook.bids
		for i := 0; i < len(ob.bids); i++ {
			if ob.bids[i] == limit {
				// remove the limit from the slice
				ob.bids = append(ob.bids[:i], ob.bids[i+1:]...)
				break
			}
		}
	} else {
		delete(ob.AskLimits, limit.Price)
		// remove limit from orderbook.asks
		for i := 0; i < len(ob.asks); i++ {
			if ob.asks[i] == limit {
				// remove the limit from the slice
				ob.asks = append(ob.asks[:i], ob.asks[i+1:]...)
				break
			}
		}

	}

}

func (ob *OrderBook) CancelOrder(order *Order) {
	limit := order.Limit
	limit.deleteOrder(order)
	if len(limit.Orders) == 0 {
		ob.clearLimit(order.Bid, limit)
	}

	delete(ob.limitOrderIdMap, order.ID)
}

func (orderBook *OrderBook) GetOrderById(id int64) *Order {
	return orderBook.limitOrderIdMap[id]
}

func (orderBook *OrderBook) GetLimitOrderIds() []int64 {
	limitOrderIds := make([]int64, 0)
	for k, _ := range orderBook.limitOrderIdMap {
		limitOrderIds = append(limitOrderIds, k)
	}
	return limitOrderIds
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
