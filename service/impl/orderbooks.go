package serviceImpl

import (
	"context"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/service"
)

type orderBookService struct {
	engine *core.MatchingEngine
}

func NewIOrderBookService(engine *core.MatchingEngine) service.IOrderBookService {
	return &orderBookService{
		engine: engine,
	}
}

func (os orderBookService) GetSnapshot(ctx context.Context, market string) (*book.BookSnapshot, error) {
	ob, err := os.engine.GetOrderBook(market)
	if err != nil {
		return nil, err
	}
	snapshot := ob.Snapshot()
	return &snapshot, nil
}

func (os orderBookService) GetLatestPrice(ctx context.Context, market string) (float64, error) {
	ob, err := os.engine.GetOrderBook(market)
	if err != nil {
		return -1.0, err
	}
	return ob.LatestPrice(), nil
}

func (os orderBookService) GetBaseQuoteAssets(ctx context.Context, market string) (string, string, error) {
	ob, err := os.engine.GetOrderBook(market)
	if err != nil {
		return "", "", err
	}
	return ob.MarketInfo().BaseAsset, ob.MarketInfo().QuoteAsset, nil
}
