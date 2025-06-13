package ohlcv

import "time"

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

type ohlcvStatistics struct {
	RecordCount  int64
	MinOpenTime  int64
	MaxCloseTime int64
	AvgVolume    float64
	TotalVolume  float64
}
