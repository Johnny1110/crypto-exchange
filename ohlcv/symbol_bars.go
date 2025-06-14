package ohlcv

import (
	"fmt"
	"sync"
	"time"
)

type RealtimeSymbolBars struct {
	symbol       string
	intervalBars map[OHLCV_INTERVAL]*OHLCVBar
	mu           sync.RWMutex
}

func NewRealtimeSymbolBars(symbol string, initPrice float64, defaultConfigs map[OHLCV_INTERVAL]IntervalConfig) *RealtimeSymbolBars {
	intervalBars := make(map[OHLCV_INTERVAL]*OHLCVBar)
	for interval, config := range defaultConfigs {
		openTime := getBucketUnixTime(time.Now(), config.Duration)
		intervalBars[interval] = NewOhlcvBar(symbol, initPrice, openTime, config.Duration)
	}
	return &RealtimeSymbolBars{
		symbol:       symbol,
		intervalBars: intervalBars,
	}
}

func (s *RealtimeSymbolBars) GetIntervalBar(interval OHLCV_INTERVAL) (OHLCVBar, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bar, ok := s.intervalBars[interval]
	if !ok || bar == nil {
		return OHLCVBar{}, false
	}
	return *bar, true
}

func (s *RealtimeSymbolBars) UpdateBar(interval OHLCV_INTERVAL, updateFunc func(*OHLCVBar)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if bar, ok := s.intervalBars[interval]; ok {
		updateFunc(bar)
	}
}

func (s *RealtimeSymbolBars) CloseBar(interval OHLCV_INTERVAL) (OHLCVBar, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if closingBar, ok := s.intervalBars[interval]; ok {
		closingBar.IsClosed = true

		// renew a bar
		nextBarOpenTime := getNextBucketUnixTime(time.Unix(closingBar.OpenTime, 0), closingBar.Duration)
		newBar := NewOhlcvBar(closingBar.Symbol, closingBar.ClosePrice, nextBarOpenTime, closingBar.Duration)
		s.intervalBars[interval] = newBar
		return *closingBar, nil
	} else {
		return OHLCVBar{}, fmt.Errorf("interval %v does not exist", interval)
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
