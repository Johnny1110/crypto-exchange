package book

import (
	"errors"
	"fmt"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"math"
	"sync"
	"time"
)

type OrderType int

const (
	MAKE_LIMIT OrderType = iota
	TAKE_LIMIT
	MARKET
)

// Trade (Match) represents a filled trade between two orders.
type Trade struct {
	BidOrderID string
	AskOrderID string
	Price      float64
	Qty        float64
	Timestamp  time.Time
}

// String implements fmt.Stringer, returning a full snapshot of the trade.
func (t Trade) String() string {
	return fmt.Sprintf(
		"Trade{BidOrderID: %q, AskOrderID: %q, Price: %.2f, Qty: %.4f, Timestamp: %s}",
		t.BidOrderID,
		t.AskOrderID,
		t.Price,
		t.Qty,
		t.Timestamp.Format(time.RFC3339),
	)
}

// OrderBook maintains buy and sell sides, and a global index for fast order lookup.
type OrderBook struct {
	market     string
	bidSide    *BookSide
	askSide    *BookSide
	orderIndex *OrderIndex
	mu         sync.Mutex
}

// NewOrderBook creates a new OrderBook instance.
func NewOrderBook(market string) *OrderBook {
	return &OrderBook{
		market:     market,
		bidSide:    NewBookSide(true),
		askSide:    NewBookSide(false),
		orderIndex: NewOrderIndex(),
	}
}

func (ob *OrderBook) PlaceOrder(orderType OrderType, order *model.Order) ([]Trade, error) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if ob.orderIndex.OrderIdExist(order.ID) {
		return nil, errors.New("Order ID already exist")
	}

	switch orderType {
	case MAKE_LIMIT:
		return nil, ob.makeLimitOrder(order)
	case TAKE_LIMIT:
		return ob.takeLimitOrder(order)
	case MARKET:
		return ob.takeMarketOrder(order)
	default:
		return nil, errors.New("Unsupported order type")
	}
}

// MakeLimitOrder adds a new limit order to the book without attempting to match. (Maker)
func (ob *OrderBook) makeLimitOrder(order *model.Order) error {
	node := &model.OrderNode{Order: order}

	// Insert node into side
	if order.Side == model.BID {
		ob.bidSide.AddOrderNode(order.Price, node)
	} else {
		ob.askSide.AddOrderNode(order.Price, node)
	}

	// Add to index for fast lookup/cancel
	ob.orderIndex.Add(node)

	return nil
}

// CancelOrder removes an existing order from the book by its ID.
func (ob *OrderBook) CancelOrder(orderID string) error {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	// Lookup index
	side, price, node, found := ob.orderIndex.Get(orderID)
	if !found {
		return errors.New("order not found")
	}

	// Remove from book side
	if side == model.BID {
		if err := ob.bidSide.RemoveOrderNode(price, node); err != nil {
			return err
		}
	} else {
		if err := ob.askSide.RemoveOrderNode(price, node); err != nil {
			return err
		}
	}

	// Remove from index
	return ob.orderIndex.Remove(orderID)
}

// Match attempts to match an incoming order against the book and returns the resulting trades.
// Any unfilled portion of the incoming order will be added to the book. (Taker)
func (ob *OrderBook) takeLimitOrder(order *model.Order) ([]Trade, error) {
	var trades []Trade
	remainingQty := order.RemainingQty
	opposite := ob.oppositeSide(order.Side)

	// loop until order fulfilled or break by stop limit
	for remainingQty > 0 {
		bestPrice, err := opposite.BestPrice()
		if err != nil || !priceCheck(order.Side, order.Price, bestPrice) {
			// no more order or hit stop limit, just break
			break
		}

		bestNode, err := opposite.PopBest()
		if err != nil {
			break
		}

		// Determine trade qty
		tradeQty := remainingQty
		if bestNode.Order.RemainingQty < remainingQty {
			tradeQty = bestNode.Order.RemainingQty
		}

		bidOrderId, askOrderId := determineOrderId(order, bestNode.Order)

		// Record trade
		trade := Trade{
			BidOrderID: bidOrderId,
			AskOrderID: askOrderId,
			Price:      bestPrice,
			Qty:        tradeQty,
			Timestamp:  time.Now(),
		}
		trades = append(trades, trade)

		// Update qty
		bestNode.Order.RemainingQty -= tradeQty
		remainingQty -= tradeQty

		// If counter-party still has leftover, put it back into book side (price level head)
		if bestNode.Order.RemainingQty > 0 {
			opposite.PutToHead(bestPrice, bestNode)
		}
	}

	// If incoming not fully filled, add remainder into book
	if remainingQty > 0 {
		order.RemainingQty = remainingQty
		err := ob.makeLimitOrder(order)
		if err != nil {
			return nil, err
		}
	}

	return trades, nil
}

func (ob *OrderBook) takeMarketOrder(order *model.Order) ([]Trade, error) {
	opposite := ob.oppositeSide(order.Side)

	if opposite.totalVolume < order.RemainingQty {
		return nil, errors.New("not enough volume")
	}
	if order.Side == model.BID {
		order.Price = math.MaxFloat64
	} else {
		order.Price = -1
	}

	return ob.takeLimitOrder(order)
}

// determineOrderId return (bidOrderId, askOrderId)
func determineOrderId(order, oppositeOrder *model.Order) (string, string) {
	if order.Side == model.BID {
		return order.ID, oppositeOrder.ID
	} else {
		return oppositeOrder.ID, order.ID
	}
}

func (ob *OrderBook) oppositeSide(side model.Side) *BookSide {
	if side == model.BID {
		return ob.askSide
	} else {
		return ob.bidSide
	}
}

func (ob *OrderBook) TotalAskVolume() float64 {
	return ob.askSide.TotalVolume()
}

func (ob *OrderBook) TotalBidVolume() float64 {
	return ob.bidSide.TotalVolume()
}

func priceCheck(orderSide model.Side, orderPrice, bestPrice float64) bool {
	if orderSide == model.BID {
		return orderPrice >= bestPrice
	} else {
		return orderPrice <= bestPrice
	}
}
