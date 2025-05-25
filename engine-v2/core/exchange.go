package core

import (
	"fmt"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"sync"
)

type Exchange struct {
	mu         sync.RWMutex
	orderbooks map[string]*book.OrderBook
}

func NewExchange(markets []string) *Exchange {
	e := &Exchange{orderbooks: make(map[string]*book.OrderBook, len(markets))}
	for _, m := range markets {
		e.orderbooks[m] = book.NewOrderBook(m)
	}
	return e
}

func (e *Exchange) GetOrderBook(market string) (*book.OrderBook, error) {
	ob, ok := e.orderbooks[market]
	if !ok {
		return nil, fmt.Errorf("market %s not found", market)
	}
	return ob, nil
}

func (e *Exchange) Markets() []string {
	markets := make([]string, 0, len(e.orderbooks))
	for m := range e.orderbooks {
		markets = append(markets, m)
	}
	return markets
}

func (e *Exchange) PlaceOrder(market string, orderType book.OrderType, order *model.Order) ([]book.Trade, error) {
	ob, err := e.GetOrderBook(market)
	if err != nil {
		return nil, err
	}
	return ob.PlaceOrder(orderType, order)
}

func (e *Exchange) CancelOrder(market string, orderID string) error {
	ob, err := e.GetOrderBook(market)
	if err != nil {
		return err
	}
	return ob.CancelOrder(orderID)
}

func (e *Exchange) Snapshot(market string) (bidPrice, bidSize, askPrice, askSize float64, err error) {
	ob, err := e.GetOrderBook(market)
	if err != nil {
		return
	}

	bidPrice, bidSize, _ = ob.BestBid()
	askPrice, askSize, _ = ob.BestAsk()
	return
}
