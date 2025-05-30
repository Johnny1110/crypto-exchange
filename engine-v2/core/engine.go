package core

import (
	"errors"
	"fmt"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/market"
	"github.com/labstack/gommon/log"
	"sync"
	"time"
)

type MatchingEngine struct {
	mu         sync.RWMutex
	orderbooks map[string]*book.OrderBook
}

func NewMatchingEngine(markets []*market.MarketInfo) (*MatchingEngine, error) {
	if len(markets) == 0 {
		return nil, errors.New("markets must have at least one market")
	}
	e := &MatchingEngine{
		orderbooks: make(map[string]*book.OrderBook, len(markets)),
	}
	for _, m := range markets {
		e.orderbooks[m.Name] = book.NewOrderBook(m)
	}
	return e, nil
}

func (e *MatchingEngine) GetOrderBook(market string) (*book.OrderBook, error) {
	ob, ok := e.orderbooks[market]
	if !ok {
		return nil, fmt.Errorf("market %s not found", market)
	}
	return ob, nil
}

func (e *MatchingEngine) ValidateMarket(market string) bool {
	_, ok := e.orderbooks[market]
	return ok
}

func (e *MatchingEngine) Markets() []string {
	markets := make([]string, 0, len(e.orderbooks))
	for m := range e.orderbooks {
		markets = append(markets, m)
	}
	return markets
}

func (e *MatchingEngine) PlaceOrder(market string, orderType book.OrderType, order *model.Order) ([]book.Trade, error) {
	ob, err := e.GetOrderBook(market)
	if err != nil {
		return nil, err
	}
	log.Infof("[Engine] PlaceOrder, market: [%s], type:[%v], mode:[%v], side:[%v] orderId:[%s], prize:[%v], size:[%v], quoteAmt:[%v]",
		market, orderType, order.Mode, order.Side, order.ID, order.Price, order.RemainingSize, order.QuoteAmount)

	return ob.PlaceOrder(orderType, order)
}

func (e *MatchingEngine) CancelOrder(market string, orderID string) (*model.Order, error) {
	ob, err := e.GetOrderBook(market)
	if err != nil {
		return nil, err
	}
	return ob.CancelOrder(orderID)
}

func (e *MatchingEngine) Snapshot(market string) (bidPrice, bidSize, askPrice, askSize float64, err error) {
	ob, err := e.GetOrderBook(market)
	if err != nil {
		return
	}

	bidPrice, bidSize, _ = ob.BestBid()
	askPrice, askSize, _ = ob.BestAsk()
	return
}

func (e *MatchingEngine) StartSnapshotRefresher() {

	ticker := time.NewTicker(300 * time.Millisecond)

	go func() {
		for range ticker.C {
			for _, market := range e.Markets() {
				ob, err := e.GetOrderBook(market)
				if err != nil {
					log.Errorf("[Engine] StartSnapshotRefresher: GetOrderBook err: %v", err)
				} else {
					ob.RefreshSnapshot()
				}
			}
		}
	}()
}
