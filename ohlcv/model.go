package ohlcv

import "time"

type OHLCV_INTERVAL string

const MIN_1 = OHLCV_INTERVAL("1m")
const MIN_15 = OHLCV_INTERVAL("15m")
const MIN_30 = OHLCV_INTERVAL("30m")
const H_1 = OHLCV_INTERVAL("1h")
const H_4 = OHLCV_INTERVAL("4h")
const D_1 = OHLCV_INTERVAL("1d")
const W_1 = OHLCV_INTERVAL("1w")
const M_1 = OHLCV_INTERVAL("1m")
const Y_1 = OHLCV_INTERVAL("1y")

type Trade struct {
	Symbol    string
	Price     float64 // price limit
	Volume    float64 // dealt qty
	Timestamp time.Time
}

type GetOhlcvDataReq struct {
	Symbol    string
	Interval  string
	StartTime time.Time
	EndTime   time.Time
	Limit     int // default 500, max 1000
}

type OHLCV struct {
	S string    `json:"s"` // status, ok
	T []int64   `json:"t"` // timestamps
	O []float64 `json:"o"` // open price
	H []float64 `json:"h"` // highest price
	L []float64 `json:"l"` // lowest price
	C []float64 `json:"c"` // closed price
	V []float64 `json:"v"` // dealt volume
}

// Internal OHLCV bar structure
type OHLCVBar struct {
	Symbol      string
	Duration    time.Duration
	OpenPrice   float64
	HighPrice   float64
	LowPrice    float64
	ClosePrice  float64
	Volume      float64
	QuoteVolume float64
	OpenTime    int64
	CloseTime   int64
	TradeCount  int64
	IsClosed    bool
}

func NewOhlcvBar(symbol string, openPrice float64, openTime int64, duration time.Duration) *OHLCVBar {
	return &OHLCVBar{
		Symbol:      symbol,
		Duration:    duration,
		OpenPrice:   openPrice,
		HighPrice:   openPrice,
		LowPrice:    openPrice,
		ClosePrice:  openPrice,
		Volume:      0.0,
		QuoteVolume: 0.0,
		OpenTime:    openTime,
		CloseTime:   openTime + int64(duration.Seconds()),
		TradeCount:  0,
		IsClosed:    false,
	}
}

type ohlcvStatistics struct {
	RecordCount  int64
	MinOpenTime  int64
	MaxCloseTime int64
	AvgVolume    float64
	TotalVolume  float64
}
