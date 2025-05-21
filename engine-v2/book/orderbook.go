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
	LIMIT OrderType = iota
	MARKET
)

// Trade (Match) represents a filled trade between two orders.
type Trade struct {
	BidOrderID string
	AskOrderID string
	Price      float64
	Size       float64
	Timestamp  time.Time
}

// String implements fmt.Stringer, returning a full snapshot of the trade.
func (t Trade) String() string {
	return fmt.Sprintf(
		"Trade{BidOrderID: %q, AskOrderID: %q, Price: %.2f, Size: %.4f, Timestamp: %s}",
		t.BidOrderID,
		t.AskOrderID,
		t.Price,
		t.Size,
		t.Timestamp.Format(time.RFC3339),
	)
}

// OrderBook maintains buy and sell sides, and a global index for fast order lookup.
type OrderBook struct {
	market      string
	bidSide     *BookSide
	askSide     *BookSide
	orderIndex  *OrderIndex
	latestPrice float64
	mu          sync.RWMutex
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

// PlaceOrder place order into order book, support LIMIT/MAKER, LIMIT/TAKER and MARKET 3 kind of scenario
func (ob *OrderBook) PlaceOrder(orderType OrderType, order *model.Order) ([]Trade, error) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if ob.orderIndex.OrderIdExist(order.ID) {
		return nil, fmt.Errorf("order ID %s already exists", order.ID)
	}

	var trades []Trade
	var err error

	switch orderType {
	case LIMIT:
		// LIMIT-Maker, place order into book and return directly
		if order.Type == model.MAKER {
			err = ob.makeLimitOrder(order)
			return nil, err
		}
		// LIMIT-Taker
		trades, err = ob.takeLimitOrder(order)
	case MARKET:
		// MARKET always will be Taker
		trades, err = ob.takeMarketOrder(order)
	default:
		return nil, fmt.Errorf("unsupported order type: %v", orderType)
	}

	// if matching error (takeLimitOrder or takeMarketOrder), return directly.
	if err != nil {
		return nil, err
	}

	// update last matching price
	ob.updateLatestPrice(trades)
	return trades, nil
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
	ob.addOrderIndex(node)

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
	return ob.removeOrderIndex(orderID)
}

// Match attempts to match an incoming order against the book and returns the resulting trades.
// Any unfilled portion of the incoming order will be added to the book. (Taker)
func (ob *OrderBook) takeLimitOrder(order *model.Order) ([]Trade, error) {
	var trades []Trade
	remainingQty := order.RemainingSize
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
		if bestNode.Order.RemainingSize < remainingQty {
			tradeQty = bestNode.Order.RemainingSize
		}

		bidOrderId, askOrderId := determineOrderId(order, bestNode.Order)

		// Record trade
		trade := Trade{
			BidOrderID: bidOrderId,
			AskOrderID: askOrderId,
			Price:      bestPrice,
			Size:       tradeQty,
			Timestamp:  time.Now(),
		}
		trades = append(trades, trade)

		// Update qty
		bestNode.Order.RemainingSize -= tradeQty
		remainingQty -= tradeQty

		// If counter-party still has leftover, put it back into book side (price level head)
		if bestNode.Order.RemainingSize > 0 {
			opposite.PutToHead(bestPrice, bestNode)
		} else {
			// If counter-party still has no leftover, remove it from orderIndex
			ob.removeOrderIndex(bestNode.Order.ID)
		}
	}

	// If incoming not fully filled, add remainder into book
	if remainingQty > 0 {
		order.RemainingSize = remainingQty
		err := ob.makeLimitOrder(order)
		if err != nil {
			return nil, err
		}
	}

	return trades, nil
}

func (ob *OrderBook) takeMarketOrder(order *model.Order) ([]Trade, error) {
	opposite := ob.oppositeSide(order.Side)

	if opposite.totalVolume < order.RemainingSize {
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
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.askSide.TotalVolume()
}

func (ob *OrderBook) TotalBidVolume() float64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.bidSide.TotalVolume()
}

func (ob *OrderBook) LatestPrice() float64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.latestPrice
}

func (ob *OrderBook) updateLatestPrice(trades []Trade) {
	if len(trades) == 0 {
		return
	}
	lastTrade := trades[len(trades)-1]
	ob.latestPrice = lastTrade.Price
}

func (ob *OrderBook) removeOrderIndex(orderId string) error {
	return ob.orderIndex.Remove(orderId)
}

func (ob *OrderBook) addOrderIndex(node *model.OrderNode) {
	ob.orderIndex.Add(node)
}

func priceCheck(orderSide model.Side, orderPrice, bestPrice float64) bool {
	if orderSide == model.BID {
		return orderPrice >= bestPrice
	} else {
		return orderPrice <= bestPrice
	}
}
