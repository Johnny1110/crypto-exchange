package ohlcv

import (
	"fmt"
	"time"
)

type OHLCV_INTERVAL string

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
		CloseTime:   openTime + int64(duration/time.Second) - 1,
		TradeCount:  0,
		IsClosed:    false,
	}
}

func (o OHLCVBar) String() string {
	openTime := time.Unix(o.OpenTime, 0)
	closeTime := time.Unix(o.CloseTime, 0)

	return fmt.Sprintf(
		"OHLCVBar{Symbol: %s, Duration: %v, O: %.4f, H: %.4f, L: %.4f, C: %.4f, V: %.2f, QV: %.2f, OpenTime: %s, CloseTime: %s, Trades: %d, Closed: %t}",
		o.Symbol,
		o.Duration,
		o.OpenPrice,
		o.HighPrice,
		o.LowPrice,
		o.ClosePrice,
		o.Volume,
		o.QuoteVolume,
		openTime.Format("2006-01-02 15:04:05"),
		closeTime.Format("2006-01-02 15:04:05"),
		o.TradeCount,
		o.IsClosed,
	)
}

func (b *OHLCVBar) Update(trade *Trade) error {
	if trade == nil {
		return fmt.Errorf("trade cannot be nil")
	}
	if trade.Price <= 0 {
		return fmt.Errorf("invalid price: %f", trade.Price)
	}
	if trade.Volume <= 0 {
		return fmt.Errorf("invalid volume: %f", trade.Volume)
	}
	// Update high (h)
	b.HighPrice = max(b.HighPrice, trade.Price)
	// update low (l)
	b.LowPrice = min(b.LowPrice, trade.Price)
	// update close (c)
	b.ClosePrice = trade.Price
	// Update volume (v)
	b.Volume += trade.Volume
	b.QuoteVolume += trade.Volume * trade.Price
	b.TradeCount++

	return nil
}

func (b *OHLCVBar) BatchUpdate(list []*Trade) {
	for _, trade := range list {
		if err := b.Update(trade); err != nil {
			fmt.Errorf("[OHLCVBar] BatchUpdate trade error: %s", err.Error())
		}
	}
}

type ohlcvStatistics struct {
	RecordCount  int64
	MinOpenTime  int64
	MaxCloseTime int64
	AvgVolume    float64
	TotalVolume  float64
}
