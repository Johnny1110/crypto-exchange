package ohlcv

import (
	"context"
)

type SimpleTradeStream struct {
	tradeCh          chan *Trade
	subscribeSymbols []string
}

func NewSimpleTradeStream(chSize int64) TradeStream {
	return &SimpleTradeStream{
		tradeCh: make(chan *Trade, chSize),
	}
}

func (s *SimpleTradeStream) Subscribe(ctx context.Context, symbols []string) (<-chan *Trade, error) {
	s.subscribeSymbols = symbols
	return s.tradeCh, nil
}

func (s *SimpleTradeStream) Close() error {
	return nil
}

func (s *SimpleTradeStream) SyncTrade(trade *Trade) {
	s.tradeCh <- trade
}
