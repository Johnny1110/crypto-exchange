package ohlcv

import (
	"context"
	"fmt"
	"github.com/labstack/gommon/log"
	"sync"
	"time"
)

type RealtimeSymbolBars struct {
	symbol       string
	intervalBars map[OHLCV_INTERVAL]map[int64]*OHLCVBar
	mu           sync.RWMutex
}

func NewRealtimeSymbolBars(symbol string, initPrice float64, defaultConfigs map[OHLCV_INTERVAL]IntervalConfig) *RealtimeSymbolBars {
	intervalBars := make(map[OHLCV_INTERVAL]map[int64]*OHLCVBar)
	for interval, config := range defaultConfigs {
		intervalBars[interval] = make(map[int64]*OHLCVBar)
		openTime := getBucketUnixTime(time.Now(), config.Duration)
		intervalBars[interval][openTime] = NewOhlcvBar(symbol, initPrice, openTime, config.Duration)
	}
	return &RealtimeSymbolBars{
		symbol:       symbol,
		intervalBars: intervalBars,
	}
}

func (s *RealtimeSymbolBars) GetIntervalBar(interval OHLCV_INTERVAL) (OHLCVBar, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucketTime := getBucketUnixTime(time.Now(), SupportedIntervals[interval].Duration)

	bar, ok := s.intervalBars[interval][bucketTime]
	if !ok || bar == nil {
		return OHLCVBar{}, false
	}
	return *bar, true
}

func (s *RealtimeSymbolBars) CloseBars(interval OHLCV_INTERVAL, targetOpenTime int64) ([]OHLCVBar, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	closeResults := make([]OHLCVBar, 0)

	if closingBars, ok := s.intervalBars[interval]; ok {
		var latestClosingBar *OHLCVBar
		for openTime, closingBar := range closingBars {
			if openTime > targetOpenTime {
				continue
			}
			latestClosingBar = latest(latestClosingBar, closingBar)
			closingBar.IsClosed = true
			closeResults = append(closeResults, *closingBar)
			delete(s.intervalBars[interval], openTime)
		}

		// renew a bar when intervalBars[interval] is empty
		if len(s.intervalBars[interval]) == 0 && latestClosingBar != nil {
			nextBarOpenTime := getNextBucketUnixTime(time.Unix(latestClosingBar.OpenTime, 0), latestClosingBar.Duration)
			newBar := NewOhlcvBar(s.symbol, latestClosingBar.ClosePrice, nextBarOpenTime, latestClosingBar.Duration)
			s.intervalBars[interval][nextBarOpenTime] = newBar
		}

		return closeResults, nil
	} else {
		return []OHLCVBar{}, fmt.Errorf("interval %v does not exist", interval)
	}
}

func (s *RealtimeSymbolBars) HasInterval(interval OHLCV_INTERVAL) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.intervalBars[interval]
	return ok
}

func (s *RealtimeSymbolBars) GetAllIntervals() []OHLCV_INTERVAL {
	s.mu.RLock()
	defer s.mu.RUnlock()
	intervals := make([]OHLCV_INTERVAL, 0, len(s.intervalBars))
	for interval := range s.intervalBars {
		intervals = append(intervals, interval)
	}
	return intervals
}

func (s *RealtimeSymbolBars) UpdateByTrades(ctx context.Context, trades []*Trade) {
	log.Debugf("[RealtimeSymbolBars] UpdateByTrades, trades: %v", trades)

	if len(trades) == 0 {
		return
	}

	type bucketKey struct {
		interval   OHLCV_INTERVAL
		bucketTime int64
	}

	bucketTrades := make(map[bucketKey][]*Trade)

	// group trades by bucketKey
	for interval, config := range SupportedIntervals {
		duration := config.Duration
		for _, trade := range trades {
			bucketTime := getBucketUnixTime(trade.Timestamp, duration)
			key := bucketKey{interval: interval, bucketTime: bucketTime}
			bucketTrades[key] = append(bucketTrades[key], trade)
		}
	}

	s.mu.Lock()

	// Batch handle
	for key, tradeList := range bucketTrades {
		intervalBars := s.intervalBars[key.interval]

		if bar, exists := intervalBars[key.bucketTime]; exists {
			bar.BatchUpdate(tradeList)
		} else {
			firstTrade := tradeList[0]
			newBar := NewOhlcvBar(s.symbol, firstTrade.Price, key.bucketTime, SupportedIntervals[key.interval].Duration)
			intervalBars[key.bucketTime] = newBar
			newBar.BatchUpdate(tradeList)
		}
	}

	s.mu.Unlock()
}

func latest(bar1 *OHLCVBar, bar2 *OHLCVBar) *OHLCVBar {
	if bar1 == nil && bar2 == nil {
		return nil
	}

	if bar1 == nil {
		return bar2
	}

	if bar2 == nil {
		return bar1
	}

	if bar1.OpenTime > bar2.OpenTime {
		return bar1
	} else {
		return bar2
	}
}
