package ohlcv

import (
	"context"
	"time"
)

type TradeStream interface {
	Subscribe(ctx context.Context, symbols []string) (<-chan *Trade, error)
	Close() error
}

// ==================== Repository ====================
type OHLCVRepository interface {
	SaveOHLCVBar(ctx context.Context, bar *ohlcvBar, interval string) error
	GetOHLCVData(ctx context.Context, req *GetOhlcvDataReq) (*OHLCV, error)
	UpdateRealtimeOHLCV(ctx context.Context, bar *ohlcvBar, interval string) error
	GetRealtimeOHLCV(ctx context.Context, symbol, interval string, openTime int64) (*ohlcvBar, error)
	UpdateStatistics(ctx context.Context, symbol, interval string, date time.Time, stats *ohlcvStatistics) error
}
