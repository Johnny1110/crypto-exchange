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
	SaveOHLCVBar(ctx context.Context, bar *OHLCVBar, interval OHLCV_INTERVAL) error
	GetOHLCVData(ctx context.Context, req *GetOhlcvDataReq) (*OHLCV, error)
	UpdateRealtimeOHLCV(ctx context.Context, bar OHLCVBar, interval OHLCV_INTERVAL) error
	GetRealtimeOHLCV(ctx context.Context, symbol, interval OHLCV_INTERVAL, openTime int64) (*OHLCVBar, error)
	UpdateStatistics(ctx context.Context, symbol, interval OHLCV_INTERVAL, date time.Time, stats *ohlcvStatistics) error
	UpsertOHLCVBars(ctx context.Context, ohlcvBars []OHLCVBar, interval OHLCV_INTERVAL) error
}
