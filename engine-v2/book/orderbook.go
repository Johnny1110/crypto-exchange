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

// Errors
var (
	ErrOrderExists          = errors.New("order already exists")
	ErrOrderNotFound        = errors.New("order not found")
	ErrInsufficientVolume   = errors.New("insufficient volume")
	ErrUnsupportedOrderType = errors.New("unsupported order type")
)

// Trade (Match) represents a filled trade between two orders.
type Trade struct {
	Market     string
	BidOrderID string
	AskOrderID string
	BidUserID  string
	AskUserID  string
	BidFeeRate float64 // Bid order fee rate
	AskFeeRate float64 // Ask order fee rate
	Price      float64 // price limit
	Size       float64 // dealt qty
	TradeValue float64 // Price * Size
	Timestamp  time.Time
}

func (t Trade) String() string {
	return fmt.Sprintf(
		"Trade{Market: %q, BidOrderID: %q, AskOrderID: %q, Price: %.8f, Size: %.8f, Value: %.8f, "+
			"BidFee: %.4f%%, AskFee: %.4f%%, Timestamp: %s}",
		t.Market, t.BidOrderID, t.AskOrderID, t.Price, t.Size, t.TradeValue,
		t.BidFeeRate*100, t.AskFeeRate*100, t.Timestamp.Format(time.RFC3339),
	)
}

// GetOrderIDBySide returns order ID by side
func (t Trade) GetOrderIDBySide(side model.Side) string {
	switch side {
	case model.BID:
		return t.BidOrderID
	case model.ASK:
		return t.AskOrderID
	}
	panic(fmt.Sprintf("invalid side: %v", side))
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

// BookSnapshot holds the top 20 bid and ask levels
type BookSnapshot struct {
	// key: priceLevel value: volume
	BidSide      []*PriceVolumePair `json:"bid_side"`
	AskSide      []*PriceVolumePair `json:"ask_side"`
	LatestPrice  float64            `json:"latest_price"`
	BestBidPrice float64            `json:"best_bid_price"`
	BestAskPrice float64            `json:"best_ask_price"`
	TotalBidSize float64            `json:"total_bid_size"`
	TotalAskSize float64            `json:"total_ask_size"`
	Timestamp    time.Time          `json:"-"`
}

func NewBookSnapshot() *BookSnapshot {
	return &BookSnapshot{
		BidSide:   make([]*PriceVolumePair, 0, 20),
		AskSide:   make([]*PriceVolumePair, 0, 20),
		Timestamp: time.Now(),
	}
}

// OrderBook maintains buy and sell sides, and a global index for fast order lookup.
type OrderBook struct {
	market      *market.MarketInfo
	bidSide     *BookSide
	askSide     *BookSide
	orderIndex  *OrderIndex
	latestPrice float64

	snapshot *BookSnapshot // best top 20 price snapshot

	// lock
	obMu       sync.RWMutex // OrderBook RW mutex
	snapshotMu sync.RWMutex // BookSnapshot RW mutex
}

// NewOrderBook creates a new OrderBook instance.
func NewOrderBook(marketInfo *market.MarketInfo) *OrderBook {
	if marketInfo == nil {
		panic("market info cannot be nil")
	}
	return &OrderBook{
		market:     marketInfo,
		bidSide:    NewBookSide(true),
		askSide:    NewBookSide(false),
		orderIndex: NewOrderIndex(),
		snapshot:   NewBookSnapshot(),
	}
}

// getSide returns the book side for the given order side
func (ob *OrderBook) getSide(side model.Side) *BookSide {
	if side == model.BID {
		return ob.bidSide
	}
	return ob.askSide
}

// getOppositeSide returns the opposite book side
func (ob *OrderBook) getOppositeSide(side model.Side) *BookSide {
	if side == model.BID {
		return ob.askSide
	}
	return ob.bidSide
}

// Snapshot returns a copy of the current book snapshot
func (ob *OrderBook) Snapshot() BookSnapshot {
	ob.snapshotMu.RLock()
	defer ob.snapshotMu.RUnlock()

	return BookSnapshot{
		BidSide:      ob.copyPriceVolumePairs(ob.snapshot.BidSide),
		AskSide:      ob.copyPriceVolumePairs(ob.snapshot.AskSide),
		BestAskPrice: ob.snapshot.BestAskPrice,
		BestBidPrice: ob.snapshot.BestBidPrice,
		LatestPrice:  ob.snapshot.LatestPrice,
		TotalAskSize: ob.snapshot.TotalAskSize,
		TotalBidSize: ob.snapshot.TotalBidSize,
		Timestamp:    ob.snapshot.Timestamp,
	}
}

// copyPriceVolumePairs creates a deep copy of price-volume pairs
func (ob *OrderBook) copyPriceVolumePairs(pairs []*PriceVolumePair) []*PriceVolumePair {
	if pairs == nil {
		return nil
	}

	copied := make([]*PriceVolumePair, len(pairs))
	for i, pair := range pairs {
		copied[i] = &PriceVolumePair{Price: pair.Price, Volume: pair.Volume}
	}
	return copied
}

// RefreshSnapshot updates the snapshot with current top 20 levels
func (ob *OrderBook) RefreshSnapshot() {
	ob.obMu.RLock()
	defer ob.obMu.RUnlock()

	ob.snapshotMu.Lock()
	defer ob.snapshotMu.Unlock()

	ob.refreshBidSnapshot()
	ob.refreshAskSnapshot()
	ob.snapshot.LatestPrice = ob.latestPrice
	bestBIdPrice, _ := ob.bidSide.BestPrice()
	bestAskPrice, _ := ob.askSide.BestPrice()
	ob.snapshot.BestBidPrice = bestBIdPrice
	ob.snapshot.BestAskPrice = bestAskPrice
	ob.snapshot.Timestamp = time.Now()
	ob.snapshot.TotalBidSize = ob.bidSide.totalVolume
	ob.snapshot.TotalAskSize = ob.askSide.totalVolume
}

// refreshBidSnapshot refreshes bid side snapshot (top 20 highest prices)
func (ob *OrderBook) refreshBidSnapshot() {
	ob.snapshot.BidSide = ob.snapshot.BidSide[:0] // Reset slice

	it := ob.bidSide.priceLevels.Iterator()
	it.End() // Start from highest price

	count := 0
	for it.Prev() && count < 20 {
		price := it.Key().(float64)
		deque := it.Value().(*util.OrderNodeDeque)
		volume := deque.Volume()

		ob.snapshot.BidSide = append(ob.snapshot.BidSide,
			NewPriceVolumePair(price, volume))
		count++
	}
}

// refreshAskSnapshot refreshes ask side snapshot (top 20 lowest prices)
func (ob *OrderBook) refreshAskSnapshot() {
	ob.snapshot.AskSide = ob.snapshot.AskSide[:0] // Reset slice

	it := ob.askSide.priceLevels.Iterator()
	it.Begin() // Start from lowest price

	count := 0
	for it.Next() && count < 20 {
		price := it.Key().(float64)
		deque := it.Value().(*util.OrderNodeDeque)
		volume := deque.Volume()

		ob.snapshot.AskSide = append(ob.snapshot.AskSide,
			NewPriceVolumePair(price, volume))
		count++
	}
}

// Order Section >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// CancelOrder removes an existing order from the book by its ID.
func (ob *OrderBook) CancelOrder(orderID string) (*model.Order, error) {
	ob.obMu.Lock()
	defer ob.obMu.Unlock()

	// Lookup index
	side, price, node, found := ob.orderIndex.Get(orderID)
	if !found {
		return nil, fmt.Errorf("%w: %s", ErrOrderNotFound, orderID)
	}

	// Remove from appropriate book side
	bookSide := ob.getSide(side)
	if err := bookSide.RemoveOrderNode(price, node); err != nil {
		return nil, fmt.Errorf("failed to remove order from book side: %w", err)
	}

	// Remove from index
	return ob.removeOrderIndex(orderID)
}

// PlaceOrder place order into order book, support LIMIT/MAKER, LIMIT/TAKER and MARKET 3 kind of scenario
func (ob *OrderBook) PlaceOrder(orderType model.OrderType, order *model.Order) ([]Trade, error) {
	if order == nil {
		return nil, errors.New("order cannot be nil")
	}

	ob.obMu.Lock()
	defer ob.obMu.Unlock()

	// Check if order ID already exists
	if ob.orderIndex.OrderIdExist(order.ID) {
		return nil, fmt.Errorf("%w: %s", ErrOrderExists, order.ID)
	}

	log.Infof("[OrderBook] PlaceOrder %s order, orderID: %s, side: %s", orderType, order.ID, order.Side)

	switch orderType {
	case model.LIMIT:
		return ob.placeLimitOrder(order)
	case model.MARKET:
		return ob.placeMarketOrder(order)
	default:
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedOrderType, orderType)
	}
}

// placeLimitOrder handles limit order placement
func (ob *OrderBook) placeLimitOrder(order *model.Order) ([]Trade, error) {
	if order.Mode == model.MAKER {
		// Maker order: add to book without matching
		err := ob.makeLimitOrder(order)
		return nil, err
	}

	// Taker order: try to match first, then add remainder to book
	trades, err := ob.takeLimitOrder(order)
	if err != nil {
		return nil, err
	}

	ob.updateLatestPrice(trades)
	return trades, nil
}

// placeMarketOrder handles market order placement
func (ob *OrderBook) placeMarketOrder(order *model.Order) ([]Trade, error) {
	var trades []Trade
	var err error

	switch order.Side {
	case model.BID:
		trades, err = ob.takeMarketBidOrder(order)
	case model.ASK:
		trades, err = ob.takeMarketAskOrder(order)
	default:
		return nil, fmt.Errorf("invalid order side: %v", order.Side)
	}

	if err != nil {
		return nil, err
	}

	ob.updateLatestPrice(trades)
	return trades, nil
}

// MakeLimitOrder adds a new limit order to the book without attempting to match. (Maker)
func (ob *OrderBook) makeLimitOrder(order *model.Order) error {
	node := &model.OrderNode{Order: order}

	// Add to appropriate side
	side := ob.getSide(order.Side)
	side.AddOrderNode(order.Price, node)
	// Add to index for fast lookup/cancel
	ob.addOrderIndex(node)

	return nil
}

// takeLimitOrder matches a limit order against the book (Taker).
// Attempts to match an incoming order against the book and returns the resulting trades.
// Any unfilled portion of the incoming order will be added to the book.
func (ob *OrderBook) takeLimitOrder(order *model.Order) ([]Trade, error) {
	var trades []Trade
	opposite := ob.getOppositeSide(order.Side)

	// Keep matching until order is filled or no more matches possible
	for order.RemainingSize > 0 {
		bestPrice, err := opposite.BestPrice()
		if err != nil || !ob.canMatch(order.Side, order.Price, bestPrice) {
			break // no more order or hit stop limit, just break
		}

		trade, shouldContinue, err := ob.executeMatch(order, opposite, bestPrice)
		if err != nil {
			return trades, err
		}
		trades = append(trades, trade)

		if !shouldContinue {
			break
		}

	}

	// Add remaining quantity to book if any
	if order.RemainingSize > 0 {
		if err := ob.makeLimitOrder(order); err != nil {
			return trades, err
		}
	}

	return trades, nil
}

// Market Order Logic Section >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
func (ob *OrderBook) takeMarketAskOrder(order *model.Order) (trades []Trade, err error) {
	opposite := ob.getOppositeSide(order.Side)

	// Check if there's enough volume
	if opposite.totalVolume < order.RemainingSize {
		return nil, fmt.Errorf("%w for market ask order in %s", ErrInsufficientVolume, ob.market.Name)
	}

	// loop until order fulfilled or break by stop limit
	for order.RemainingSize > 0 {
		bestNode, err := opposite.PopBest()
		if err != nil {
			log.Errorf("[OrderBook] critical error in market ask order %s: %v", order.ID, err)
			break
		}

		// Determine trade qty
		tradeQty := min(order.RemainingSize, bestNode.Order.RemainingSize)
		trade := ob.createTrade(order, bestNode.Order, bestNode.Order.Price, tradeQty)
		trades = append(trades, trade)

		// Update qty
		bestNode.Order.RemainingSize -= tradeQty
		order.RemainingSize -= tradeQty

		// Handle counter-party
		if bestNode.Order.RemainingSize > 0 {
			opposite.PutToHead(bestNode.Order.Price, bestNode)
		} else {
			if _, err := ob.orderIndex.Remove(bestNode.Order.ID); err != nil {
				log.Errorf("[OrderBook] failed to remove order from index: %s", bestNode.Order.ID)
			}
		}

	}

	return trades, err
}

func (ob *OrderBook) takeMarketBidOrder(order *model.Order) (trades []Trade, err error) {
	opposite := ob.getOppositeSide(order.Side)

	// Check if there's enough quote amount
	if opposite.totalQuoteAmount < order.QuoteAmount {
		return nil, fmt.Errorf("%w for market bid order in %s",
			ErrInsufficientVolume, ob.market.Name)
	}

	remainingQuoteAmt := order.QuoteAmount

	// Consume all remainingQuoteAmt
	for remainingQuoteAmt > 0 {
		bestNode, err := opposite.PopBest()
		if err != nil {
			log.Errorf("[OrderBook] critical error in market bid order %s: %v", order.ID, err)
			break
		}

		oppositeOrder := bestNode.Order
		oppositeOrderQuoteAmt := oppositeOrder.RemainingSize * oppositeOrder.Price

		// Determine trade qty
		var tradeQty float64
		if remainingQuoteAmt >= oppositeOrderQuoteAmt {
			// eat all oppositeOrder qty
			tradeQty = oppositeOrder.RemainingSize
		} else {
			tradeQty = remainingQuoteAmt / oppositeOrder.Price
		}

		trade := ob.createTrade(order, oppositeOrder, oppositeOrder.Price, tradeQty)
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
			if _, err := ob.removeOrderIndex(oppositeOrder.ID); err != nil {
				log.Errorf("[OrderBook] failed to remove order from index: %s", oppositeOrder.ID)
			}
		}
	}

	return trades, err
}

// Market Order Logic Section <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

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

// canMatch checks if an order can match at the given price
func (ob *OrderBook) canMatch(orderSide model.Side, orderPrice, bestPrice float64) bool {
	if orderSide == model.BID {
		return orderPrice >= bestPrice
	}
	return orderPrice <= bestPrice
}

// executeMatch executes a single match between orders
func (ob *OrderBook) executeMatch(order *model.Order, opposite *BookSide, bestPrice float64) (Trade, bool, error) {
	bestNode, err := opposite.PopBest()
	if err != nil {
		return Trade{}, false, err
	}

	// Calculate trade quantity
	tradeQty := min(order.RemainingSize, bestNode.Order.RemainingSize)

	// Create trade
	trade := ob.createTrade(order, bestNode.Order, bestPrice, tradeQty)

	// Update remaining quantities
	bestNode.Order.RemainingSize -= tradeQty
	order.RemainingSize -= tradeQty

	// Handle counter-party order
	if bestNode.Order.RemainingSize > 0 {
		// Put back to front of price level
		opposite.PutToHead(bestPrice, bestNode)
	} else {
		// Remove from index
		if _, err := ob.orderIndex.Remove(bestNode.Order.ID); err != nil {
			log.Errorf("[OrderBook] failed to remove order from index: %s", bestNode.Order.ID)
		}
	}

	return trade, true, nil
}

// createTrade creates a trade record
func (ob *OrderBook) createTrade(order1, order2 *model.Order, price, size float64) Trade {
	bidOrder, askOrder := ob.determineTradeOrders(order1, order2)

	return Trade{
		Market:     ob.market.Name,
		BidOrderID: bidOrder.ID,
		AskOrderID: askOrder.ID,
		BidUserID:  bidOrder.UserID,
		AskUserID:  askOrder.UserID,
		BidFeeRate: bidOrder.FeeRate,
		AskFeeRate: askOrder.FeeRate,
		Price:      price,
		Size:       size,
		Timestamp:  time.Now(),
	}
}

// determineTradeOrders determines bid/ask order IDs and user IDs
func (ob *OrderBook) determineTradeOrders(order1, order2 *model.Order) (bidOrder *model.Order, askOrder *model.Order) {
	if order1.Side == model.BID {
		return order1, order2
	}
	return order2, order1
}

// PutOrder put order into OrderBook directly
func (ob *OrderBook) PutOrder(order *model.Order) error {
	ob.obMu.Lock()
	defer ob.obMu.Unlock()

	err := ob.makeLimitOrder(order)
	if err != nil {
		return err
	}
	return nil
}

func (ob *OrderBook) UpdateLatestPrice(price float64) {
	ob.obMu.Lock()
	defer ob.obMu.Unlock()

	ob.latestPrice = price
}
