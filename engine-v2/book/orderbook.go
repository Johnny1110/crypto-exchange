package book

import (
	"errors"
	"fmt"
	"github.com/johnny1110/crypto-exchange/engine-v2/market"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/engine-v2/util"
	"github.com/labstack/gommon/log"
	"sync"
	"time"
)

// Trade (Match) represents a filled trade between two orders.
type Trade struct {
	BidOrderID string
	AskOrderID string
	BidUserID  string
	AskUserID  string
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

func (t Trade) GeOrderIDBySide(side model.Side) string {
	switch side {
	case model.BID:
		return t.BidOrderID
	case model.ASK:
		return t.AskOrderID
	}
	panic("unreachable")
}

// BookSnapshot only hold bid highest 20, and ask lowest 20.
type BookSnapshot struct {
	// key: priceLevel value: volume
	BidSide     []*PriceVolumePair
	AskSide     []*PriceVolumePair
	LatestPrice float64
}

type PriceVolumePair struct {
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
}

func NewPriceVolumePair(price float64, volume float64) *PriceVolumePair {
	return &PriceVolumePair{
		Price:  price,
		Volume: volume,
	}
}

func NewBookSnapshot() *BookSnapshot {
	return &BookSnapshot{
		BidSide:     make([]*PriceVolumePair, 0, 20),
		AskSide:     make([]*PriceVolumePair, 0, 20),
		LatestPrice: 0.0,
	}
}

// OrderBook maintains buy and sell sides, and a global index for fast order lookup.
type OrderBook struct {
	market      *market.MarketInfo
	bidSide     *BookSide
	askSide     *BookSide
	orderIndex  *OrderIndex
	latestPrice float64
	obMu        sync.RWMutex  // OrderBook RW mutex
	snapshot    *BookSnapshot // best top 20 price snapshot
	snapshotMu  sync.RWMutex  // BookSnapshot RW mutex
}

// NewOrderBook creates a new OrderBook instance.
func NewOrderBook(market *market.MarketInfo) *OrderBook {
	return &OrderBook{
		market:     market,
		bidSide:    NewBookSide(true),
		askSide:    NewBookSide(false),
		orderIndex: NewOrderIndex(),
		snapshot:   NewBookSnapshot(),
	}
}

// Snapshot return snapshot
func (ob *OrderBook) Snapshot() BookSnapshot {
	ob.snapshotMu.Lock()
	defer ob.snapshotMu.Unlock()
	bidCopy := make([]*PriceVolumePair, len(ob.snapshot.BidSide))
	copy(bidCopy, ob.snapshot.BidSide)

	askCopy := make([]*PriceVolumePair, len(ob.snapshot.AskSide))
	copy(askCopy, ob.snapshot.AskSide)

	return BookSnapshot{
		BidSide:     bidCopy,
		AskSide:     askCopy,
		LatestPrice: ob.snapshot.LatestPrice,
	}
}

// Refresh Do refresh snapshot, read lock orderbook, and write lock snapshot
// Run a 500 ms job to refresh
func (ob *OrderBook) RefreshSnapshot() {
	ob.obMu.RLock()
	ob.snapshotMu.Lock()
	defer ob.snapshotMu.Unlock()
	defer ob.obMu.RUnlock()

	// clean bidSide snapshot
	ob.snapshot.BidSide = ob.snapshot.BidSide[:0]
	// iterate bidPriceLevel from max collect top 20 (price:volume) and save into ob.snapshot
	it := ob.bidSide.priceLevels.Iterator()
	it.End() // move to the largest key
	count := 0
	for it.Prev() && count < 20 {
		price := it.Key().(float64)
		deque := it.Value().(*util.OrderNodeDeque)
		volume := deque.Volume()
		ob.snapshot.BidSide = append(ob.snapshot.BidSide, NewPriceVolumePair(price, volume))
		count++
	}

	// clean askSide snapshot
	ob.snapshot.AskSide = ob.snapshot.AskSide[:0]
	// iterate askPriceLevel from min collect top 20 (price:volume) and save into ob.snapshot
	it = ob.askSide.priceLevels.Iterator()
	it.Begin() // move to smallest key
	count = 0
	for it.Next() && count < 20 {
		price := it.Key().(float64)
		deque := it.Value().(*util.OrderNodeDeque)
		volume := deque.Volume()
		ob.snapshot.AskSide = append(ob.snapshot.AskSide, NewPriceVolumePair(price, volume))
		count++
	}

	ob.snapshot.LatestPrice = ob.latestPrice
}

// PlaceOrder place order into order book, support LIMIT/MAKER, LIMIT/TAKER and MARKET 3 kind of scenario
func (ob *OrderBook) PlaceOrder(orderType model.OrderType, order *model.Order) ([]Trade, error) {
	ob.obMu.Lock()
	defer ob.obMu.Unlock()

	// check id exists
	if ob.orderIndex.OrderIdExist(order.ID) {
		return nil, fmt.Errorf("order ID %s already exists", order.ID)
	}

	var trades []Trade
	var err error

	switch orderType {
	case model.LIMIT:
		// LIMIT-Maker, place order into book and return directly
		if order.Mode == model.MAKER {
			log.Infof("[OrderBook] PlaceOrder (maker) LIMIT order, orderID:[%s]", order.ID)
			err = ob.makeLimitOrder(order)
			return nil, err
		} else {
			log.Infof("[OrderBook] PlaceOrder (taker) LIMIT order, orderID:[%s]", order.ID)
			// LIMIT-Taker
			trades, err = ob.takeLimitOrder(order)
		}
		break
	case model.MARKET:
		// MARKET always will be Taker
		log.Infof("[OrderBook] PlaceOrder (taker) MARKET order, orderID:[%s]", order.ID)
		if order.Side == model.BID {
			// market bid eat opposite order based on order.quoteAmount
			trades, err = ob.takeMarketBidOrder(order)
		} else {
			// market ask eat opposite order based on order.RemainingSize, just like limit ask did.
			trades, err = ob.takeMarketAskOrder(order)
		}
		break
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
func (ob *OrderBook) CancelOrder(orderID string) (*model.Order, error) {
	ob.obMu.Lock()
	defer ob.obMu.Unlock()

	// Lookup index
	side, price, node, found := ob.orderIndex.Get(orderID)
	if !found {
		return nil, errors.New("order not found")
	}

	// Remove from book side
	if side == model.BID {
		if err := ob.bidSide.RemoveOrderNode(price, node); err != nil {
			return nil, err
		}
	} else {
		if err := ob.askSide.RemoveOrderNode(price, node); err != nil {
			return nil, err
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

		bidOrderId, bidUserId, askOrderId, askUserId := determineOrderId(order, bestNode.Order)

		// Record trade
		trade := Trade{
			BidOrderID: bidOrderId,
			AskOrderID: askOrderId,
			BidUserID:  bidUserId,
			AskUserID:  askUserId,
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
			// If counter-party has no leftover, remove it from orderIndex
			_, err := ob.removeOrderIndex(bestNode.Order.ID)
			if err != nil {
				log.Errorf("[OrderBook] takeLimitOrder critical error, failed to remove orderIdx ID:[%s]", bestNode.Order.ID)
			}
		}
	}

	order.RemainingSize = remainingQty

	// If incoming not fully filled, add remainder into book
	if order.RemainingSize > 0 {
		err := ob.makeLimitOrder(order)
		if err != nil {
			return nil, err
		}
	}

	return trades, nil
}

// Market Order Logic Section >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
func (ob *OrderBook) takeMarketAskOrder(order *model.Order) (trades []Trade, err error) {
	opposite := ob.oppositeSide(order.Side)
	if opposite.totalVolume < order.RemainingSize {
		log.Warnf("[OrderBook] takeMarketAskOrder failed, no enough volume sit in [%s] market", ob.market.Name)
		return nil, errors.New("not enough volume for market ask order")
	}

	// loop until order fulfilled or break by stop limit
	for order.RemainingSize > 0 {
		bestNode, err := opposite.PopBest()
		if err != nil {
			log.Errorf("[OrderBook] takeMarketAskOrder critical error, "+
				"not enough volume for market order while matching order: [%s]", order.ID)
			break
		}

		oppositeOrder := bestNode.Order

		// Determine trade qty
		tradeQty := order.RemainingSize
		if oppositeOrder.RemainingSize < tradeQty {
			tradeQty = oppositeOrder.RemainingSize
		}

		bidOrderId, bidUserId, askOrderId, askUserId := determineOrderId(order, bestNode.Order)

		// Record trade
		trade := Trade{
			BidOrderID: bidOrderId,
			AskOrderID: askOrderId,
			BidUserID:  bidUserId,
			AskUserID:  askUserId,
			Price:      oppositeOrder.Price,
			Size:       tradeQty,
			Timestamp:  time.Now(),
		}
		trades = append(trades, trade)

		// Update qty
		oppositeOrder.RemainingSize -= tradeQty
		order.RemainingSize -= tradeQty

		// If counter-party still has leftover, put it back into book side (price level head)
		if oppositeOrder.RemainingSize > 0 {
			opposite.PutToHead(oppositeOrder.Price, bestNode)
		} else {
			// If counter-party has no leftover, remove it from orderIndex
			_, err := ob.removeOrderIndex(oppositeOrder.ID)
			if err != nil {
				log.Errorf("[OrderBook] takeMarketAskOrder critical error, failed to remove orderIdx ID:[%s]", oppositeOrder.ID)
			}
		}
	}

	return trades, err
}

func (ob *OrderBook) takeMarketBidOrder(order *model.Order) (trades []Trade, err error) {
	opposite := ob.oppositeSide(order.Side)
	if opposite.totalQuoteAmount < order.QuoteAmount {
		log.Warnf("[OrderBook] takeMarketBidOrder failed, no enough volume sit in [%s] market", ob.market.Name)
		return nil, errors.New("not enough volume for market bid order")
	}

	remainingQuoteAmt := order.QuoteAmount

	// Consume all remainingQuoteAmt
	for remainingQuoteAmt > 0 {
		bestNode, err := opposite.PopBest()
		if err != nil {
			log.Errorf("[OrderBook] takeMarketOrder critical error, "+
				"not enough volume for market order while matching order: [%s]", order.ID)
			break
		}

		oppositeOrder := bestNode.Order
		oppositeOrderQuoteAmt := oppositeOrder.RemainingSize * oppositeOrder.Price

		// Determine trade qty
		tradeQty := 0.0
		if remainingQuoteAmt >= oppositeOrderQuoteAmt {
			// eat all oppositeOrder qty
			tradeQty = oppositeOrder.RemainingSize
		} else {
			tradeQty = remainingQuoteAmt / oppositeOrder.Price
		}

		bidOrderId, bidUserId, askOrderId, askUserId := determineOrderId(order, bestNode.Order)

		// Record trade
		trade := Trade{
			BidOrderID: bidOrderId,
			AskOrderID: askOrderId,
			BidUserID:  bidUserId,
			AskUserID:  askUserId,
			Price:      oppositeOrder.Price,
			Size:       tradeQty,
			Timestamp:  time.Now(),
		}
		trades = append(trades, trade)

		// Update qty
		oppositeOrder.RemainingSize -= tradeQty
		remainingQuoteAmt -= tradeQty * oppositeOrder.Price
		order.OriginalSize += tradeQty // increase eaten order's OriginalSize

		// If counter-party still has leftover, put it back into book side (price level head)
		if oppositeOrder.RemainingSize > 0 {
			opposite.PutToHead(oppositeOrder.Price, bestNode)
		} else {
			// If counter-party has no leftover, remove it from orderIndex
			_, err := ob.removeOrderIndex(oppositeOrder.ID)
			if err != nil {
				log.Errorf("[OrderBook] takeMarketBidOrder critical error, failed to remove orderIdx ID:[%s]", oppositeOrder.ID)
			}
		}
	}

	return trades, err
}

// Market Order Logic Section <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

// determineOrderId return (bidOrderId, bidUserId, askOrderId, askUserId)
func determineOrderId(order, oppositeOrder *model.Order) (string, string, string, string) {
	if order.Side == model.BID {
		return order.ID, order.UserID, oppositeOrder.ID, oppositeOrder.UserID
	} else {
		return oppositeOrder.ID, oppositeOrder.UserID, order.ID, order.UserID
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
	ob.obMu.RLock()
	defer ob.obMu.RUnlock()
	return ob.askSide.TotalVolume()
}

func (ob *OrderBook) TotalAskQuoteAmount() float64 {
	ob.obMu.RLock()
	defer ob.obMu.RUnlock()
	return ob.askSide.TotalQuoteAmount()
}

func (ob *OrderBook) TotalBidVolume() float64 {
	ob.obMu.RLock()
	defer ob.obMu.RUnlock()
	return ob.bidSide.TotalVolume()
}

func (ob *OrderBook) TotalBidQuoteAmount() float64 {
	ob.obMu.RLock()
	defer ob.obMu.RUnlock()
	return ob.bidSide.TotalQuoteAmount()
}

func (ob *OrderBook) LatestPrice() float64 {
	ob.obMu.RLock()
	defer ob.obMu.RUnlock()
	return ob.latestPrice
}

func (ob *OrderBook) updateLatestPrice(trades []Trade) {
	if len(trades) == 0 {
		return
	}
	lastTrade := trades[len(trades)-1]
	ob.latestPrice = lastTrade.Price
}

func (ob *OrderBook) removeOrderIndex(orderId string) (*model.Order, error) {
	return ob.orderIndex.Remove(orderId)
}

func (ob *OrderBook) addOrderIndex(node *model.OrderNode) {
	ob.orderIndex.Add(node)
}

func (ob *OrderBook) BestBid() (float64, float64, error) {
	ob.obMu.RLock()
	defer ob.obMu.RUnlock()
	bestPrice, err := ob.bidSide.BestPrice()
	if err != nil {
		return 0, 0, err
	}
	volume := ob.bidSide.TotalVolume()
	return bestPrice, volume, nil
}

func (ob *OrderBook) BestAsk() (float64, float64, error) {
	ob.obMu.RLock()
	defer ob.obMu.RUnlock()

	bestPrice, err := ob.askSide.BestPrice()
	if err != nil {
		return 0, 0, err
	}
	volume := ob.askSide.TotalVolume()
	return bestPrice, volume, nil
}

func (ob *OrderBook) MarketInfo() *market.MarketInfo {
	return ob.market
}

func (ob *OrderBook) GetAssets() (string, string) {
	return ob.market.BaseAsset, ob.market.QuoteAsset
}

func priceCheck(orderSide model.Side, orderPrice, bestPrice float64) bool {
	if orderSide == model.BID {
		return orderPrice >= bestPrice
	} else {
		return orderPrice <= bestPrice
	}
}
