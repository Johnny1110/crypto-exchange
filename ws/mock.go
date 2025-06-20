package ws

import (
	"context"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/ohlcv"
	"time"
)

type MockDataFeeder struct{}

func Feed(ctx context.Context, pkg WSFeedPackage) {
	//TODO
}

func (m *MockDataFeeder) Start(ctx context.Context, hub *Hub) {
	ohlcvTicker := time.NewTicker(5 * time.Second)
	orderbookTicker := time.NewTicker(2 * time.Second)

	defer ohlcvTicker.Stop()
	defer orderbookTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ohlcvTicker.C:
			ohlcvData := ohlcv.OHLCVBar{
				Symbol:     "BTC-USDT",
				Duration:   1 * time.Minute,
				OpenPrice:  50000.1,
				HighPrice:  51000.3,
				LowPrice:   50000.1,
				ClosePrice: 50000.1,
				Volume:     1000,
			}

			key := SubscriptionKey{
				Channel: OHLCV,
				Params: OHLCVReqParams{
					Symbol:   "BTC-USDT",
					Interval: "1m",
				},
			}

			hub.BroadcastToSubscribers(key, ohlcvData)

		case <-orderbookTicker.C:
			orderbookData := book.BookSnapshot{
				BidSide:      make([]*book.PriceVolumePair, 0),
				AskSide:      make([]*book.PriceVolumePair, 0),
				BestBidPrice: 10.0,
				BestAskPrice: 20.0,
				LatestPrice:  15.0,
				TotalAskSize: 1,
				TotalBidSize: 1,
				Timestamp:    time.Now(),
			}

			key := SubscriptionKey{
				Channel: ORDERBOOK,
				Params: OrderBookReqParams{
					Market: "BTC-USDT",
				},
			}

			hub.BroadcastToSubscribers(key, orderbookData)
		}
	}
}
