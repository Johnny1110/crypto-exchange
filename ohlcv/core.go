package ohlcv

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/labstack/gommon/log"
	"sync"
	"time"
)

// ========================= Config =========================

type IntervalConfig struct {
	Duration time.Duration
	Table    string
}

var SupportedIntervals = map[string]IntervalConfig{
	"1h": {Duration: time.Hour, Table: "ohlcv_1h"},
	"1d": {Duration: 24 * time.Hour, Table: "ohlcv_1d"},
	"1w": {Duration: 7 * 24 * time.Hour, Table: "ohlcv_1w"},
}

// ========================= New Aggregator =========================
type AggregatorSetupConfig struct {
	BatchSize     int
	FlushInterval time.Duration
}

type OHLCVAggregator struct {
	db          *sql.DB
	repo        OHLCVRepository
	tradeStream TradeStream

	// Real-time OHLCV bars cache (symbol -> interval -> bar)
	realtimeBars sync.Map // map[string]map[string]*OHLCVBar
	timers       sync.Map // map[string]*time.Timer

	// Channels for internal communication
	tradeCh chan *Trade
	stopCh  chan struct{}
	// Timers for each interval

	// Configuration
	batchSize     int
	flushInterval time.Duration
	// Statistics tracking
	statsCache sync.Map // map[string]*OHLCVStatistics
}

func NewOHLCVAggregator(repo OHLCVRepository, stream TradeStream, config *AggregatorSetupConfig) *OHLCVAggregator {
	if config.BatchSize <= 0 {
		// min batch size = 100
		config.BatchSize = 100
	}
	if config.FlushInterval <= 0 {
		// min flush interval = 5 secs
		config.FlushInterval = 5 * time.Second
	}

	return &OHLCVAggregator{
		repo:          repo,
		tradeStream:   stream,
		tradeCh:       make(chan *Trade, 1000),
		stopCh:        make(chan struct{}),
		batchSize:     config.BatchSize,
		flushInterval: config.FlushInterval,
	}
}

// ========================= Aggregator expose func =========================

func (a *OHLCVAggregator) Start(ctx context.Context, symbols []string) error {
	// Subscribe to trade stream.
	tradeStreamCh, err := a.tradeStream.Subscribe(ctx, symbols)
	if err != nil {
		return fmt.Errorf("[OHLCVAggregator] failed to subscribe trade stream: %w", err)
	}

	// Start trade processing goroutine
	go a.processTradeStream(ctx, tradeStreamCh)
	// Start aggregation goroutine
	go a.aggregateTrades(ctx)
	// Start periodic flush
	go a.periodicFlush(ctx)
	// Start interval timers
	go a.manageIntervalTimers(ctx)

	log.Infof("[OHLCVAggregator] OHLCV aggregator started successfully")
	return nil
}

func (a *OHLCVAggregator) Stop() error {
	close(a.stopCh)
	return a.tradeStream.Close()
}

// ========================= Aggregator Domain Logic =========================

// processTradeStream receive trade from tradeStreamCh, and push data into internal channel: tradeCh
func (a *OHLCVAggregator) processTradeStream(ctx context.Context, tradeStreamCh <-chan *Trade) {
	for {
		select {
		case trade := <-tradeStreamCh:
			if trade != nil {
				select {
				case a.tradeCh <- trade:
				default:
					log.Warnf("[OHLCVAggregator] Trade channel full, dropping trade: %v", trade)
				}
			}
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		}
	}
}

// aggregateTrades listen on tradeCh, receive trade data and process data (batch)
func (a *OHLCVAggregator) aggregateTrades(ctx context.Context) {
	// create trade data batch container
	tradeBatch := make([]*Trade, 0, a.batchSize)
	ticker := time.NewTicker(a.flushInterval)
	for {
		select {
		case trade := <-a.tradeCh:
			tradeBatch = append(tradeBatch, trade)

			// Process batch when full
			if len(tradeBatch) >= a.batchSize {
				a.processTradeBatch(ctx, tradeBatch)
				tradeBatch = tradeBatch[:0] // Reset batch slice
			}

		case <-ticker.C:
			// Process remaining trades on timeout
			if len(tradeBatch) > 0 {
				a.processTradeBatch(ctx, tradeBatch)
				tradeBatch = tradeBatch[:0] // Reset batch slice
			}

		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		}
	}
}

// processTradeBatch group by symbol and do process.
func (a *OHLCVAggregator) processTradeBatch(ctx context.Context, trades []*Trade) {
	// Group trades by symbol
	symbolTrades := make(map[string][]*Trade)
	for _, trade := range trades {
		symbolTrades[trade.Symbol] = append(symbolTrades[trade.Symbol], trade)
	}

	// Process each symbol's trades
	for symbol, symbolTradeList := range symbolTrades {
		a.processSymbolTrades(ctx, symbol, symbolTradeList)
	}
}

// processSymbolTrades process trades by (symbol)
func (a *OHLCVAggregator) processSymbolTrades(ctx context.Context, symbol string, trades []*Trade) {
	// Process for each supported interval (1h, 1d, 1w...)
	for interval, config := range SupportedIntervals {
		a.processTradesForInterval(ctx, symbol, interval, config, trades)
	}
}

// processTradesForInterval process trades by (symbol)> (interval)    ex: ETH-UST, 1h
func (a *OHLCVAggregator) processTradesForInterval(ctx context.Context, symbol, interval string, config IntervalConfig, trades []*Trade) {
	// Group trades by time buckets
	buckets := make(map[int64][]*Trade)

	for _, trade := range trades {
		bucketTime := a.getBucketTime(trade.Timestamp, config.Duration)
		buckets[bucketTime] = append(buckets[bucketTime], trade)
	}

	// Process each bucket
	for bucketTime, bucketTrades := range buckets {
		a.updateOHLCVBar(ctx, symbol, interval, bucketTime, config.Duration, bucketTrades)
	}
}

// getBucketTime input tradeTime and interval return the timestamp align the interval boundary
// Example: 1 hr boundary: (1)2024-01-01 00:00:00, (2)2024-01-01 00:01:00, (3)2024-01-01 00:02:00 (4)...
func (a *OHLCVAggregator) getBucketTime(tradeTime time.Time, interval time.Duration) int64 {
	// Align timestamp to interval boundary
	tradeUnixTime := tradeTime.Unix()
	intervalSeconds := int64(interval.Seconds())
	// tradeUnixTime / intervalSeconds = bucket

	// original tradeUnixTime: 2024-01-01 14:25:30 (Unix: 1704117930)
	// intervalSeconds = 3600 (1hr)
	//
	// calculate:
	// 1704117930 / 3600 = 473365.536... -> int64 divide int64 will cut decimal places = 473365
	// 473365 * 3600 = 1704117000 (ignore mm:ss)
	//
	// Result: 2024-01-01 14:00:00 (align the 1hr boundary)
	return (tradeUnixTime / intervalSeconds) * intervalSeconds
}

func (a *OHLCVAggregator) updateOHLCVBar(ctx context.Context, symbol, interval string, openTime int64, duration time.Duration, trades []*Trade) {
	closeTime := openTime + int64(duration.Seconds()) - 1

	// Get or create realtime bar
	var bar *ohlcvBar
	if existingBar, err := a.repo.GetRealtimeOHLCV(ctx, symbol, interval, openTime); err == nil && existingBar != nil {
		bar = existingBar
	} else {
		// Create new bar with first trade
		firstTrade := trades[0]
		bar = &ohlcvBar{
			Symbol:     symbol,
			OpenPrice:  firstTrade.Price, // create (o)
			HighPrice:  firstTrade.Price,
			LowPrice:   firstTrade.Price,
			ClosePrice: firstTrade.Price,
			Volume:     0,
			OpenTime:   openTime,
			CloseTime:  closeTime,
			IsClosed:   false,
		}
	}

	// Update bar with all trades
	for _, trade := range trades {
		// update h, l, c, v
		a.updateBarWithTrade(bar, trade)
	}

	// Save/update realtime bar
	if err := a.repo.UpdateRealtimeOHLCV(ctx, bar, interval); err != nil {
		log.Errorf("[OHLCVAggregator] Failed to update realtime OHLCV: %v", err)
	}

	// Store in memory cache for quick access
	a.storeRealtimeBar(symbol, interval, bar)
}

// updateBarWithTrade update h, l, c, v
func (a *OHLCVAggregator) updateBarWithTrade(bar *ohlcvBar, trade *Trade) {
	// Update high (h)
	if trade.Price > bar.HighPrice {
		bar.HighPrice = trade.Price
	}
	// update low (l)
	if trade.Price < bar.LowPrice {
		bar.LowPrice = trade.Price
	}
	// update close (c)
	bar.ClosePrice = trade.Price

	// Update volume (v)
	bar.Volume += trade.Volume
	bar.QuoteVolume += trade.Volume * trade.Price
	bar.TradeCount++
}

func (a *OHLCVAggregator) storeRealtimeBar(symbol, interval string, bar *ohlcvBar) {
	symbolKey := symbol

	// realtimeBar: <symbol>: map<intervalBar>
	// intervalBar: <intervalType>: *ohlcvBar

	if symbolBars, ok := a.realtimeBars.Load(symbolKey); ok {
		if intervalBars, ok := symbolBars.(map[string]*ohlcvBar); ok {
			intervalBars[interval] = bar
		}
	} else {
		newSymbolBars := make(map[string]*ohlcvBar)
		newSymbolBars[interval] = bar
		a.realtimeBars.Store(symbolKey, newSymbolBars)
	}
}

// ==================== Interval Timer Management ====================

func (a *OHLCVAggregator) manageIntervalTimers(ctx context.Context) {
	for interval, config := range SupportedIntervals {
		go a.startIntervalTimer(ctx, interval, config)
	}
}

// startIntervalTimer process interval (1h 1d 1w ..), if reached close bar time, do closeIntervalBars()
func (a *OHLCVAggregator) startIntervalTimer(ctx context.Context, interval string, config IntervalConfig) {
	// Calculate next interval boundary
	now := time.Now()
	nextBoundary := a.getNextBoundary(now, config.Duration)

	timer := time.NewTimer(time.Until(nextBoundary))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			a.closeIntervalBars(ctx, interval, nextBoundary.Add(-config.Duration).Unix())
			// Set next timer
			nextBoundary = nextBoundary.Add(config.Duration)
			timer.Reset(time.Until(nextBoundary))

		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		}
	}
}

// getNextBoundary
func (a *OHLCVAggregator) getNextBoundary(current time.Time, interval time.Duration) time.Time {
	intervalSeconds := int64(interval.Seconds()) // 1r = 3600
	currentSeconds := current.Unix()             // 2024-01-01 14:25:30 (Unix: 1704117930)
	nextBoundarySeconds := ((currentSeconds / intervalSeconds) + 1) * intervalSeconds
	// (1704117930/3600 + 1) * 3600 :ps current sec boundary add 1 sec
	return time.Unix(nextBoundarySeconds, 0) // convert secs to time.Time
}

func (a *OHLCVAggregator) closeIntervalBars(ctx context.Context, interval string, openTime int64) {
	// Find and close all bars for this interval and time
	a.realtimeBars.Range(func(symbolKey, symbolBars interface{}) bool {
		if intervalBars, ok := symbolBars.(map[string]*ohlcvBar); ok {
			if bar, exists := intervalBars[interval]; exists && bar.OpenTime == openTime {
				// Mark as closed and save to main table
				bar.IsClosed = true
				if err := a.repo.SaveOHLCVBar(ctx, bar, interval); err != nil {
					log.Errorf("[OHLCVAggregator] Failed to save closed OHLCV bar: %v", err)
				}
				// Remove from realtime cache
				delete(intervalBars, interval)
			}
		}
		return true
	})
}

// ==================== Periodic Tasks ====================

func (a *OHLCVAggregator) periodicFlush(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 5) // Flush every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.flushRealtimeBars(ctx)
			a.updateDailyStatistics(ctx)
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		}
	}
}

func (a *OHLCVAggregator) flushRealtimeBars(ctx context.Context) {
	a.realtimeBars.Range(func(symbolKey, symbolBars interface{}) bool {
		if intervalBars, ok := symbolBars.(map[string]*ohlcvBar); ok {
			for interval, bar := range intervalBars {
				if err := a.repo.UpdateRealtimeOHLCV(ctx, bar, interval); err != nil {
					log.Warnf("[OHLCVAggregator] Failed to flush realtime bar: %v", err)
				}
			}
		}
		return true
	})
}

func (a *OHLCVAggregator) updateDailyStatistics(ctx context.Context) {
	// Update statistics for each symbol and interval
	//today := time.Now().Truncate(24 * time.Hour)
	//
	//a.realtimeBars.Range(func(symbolKey, symbolBars interface{}) bool {
	//	symbol := symbolKey.(string)
	//	for interval := range SupportedIntervals {
	//		stats := &ohlcvStatistics{
	//			// TODO Calculate statistics from recent data
	//		}
	//
	//		if err := a.repo.UpdateStatistics(ctx, symbol, interval, today, stats); err != nil {
	//			log.Errorf("[OHLCVAggregator] Failed to update statistics: %v", err)
	//		}
	//	}
	//	return true
	//})
}

// ==================== Public Query Methods ====================

func (a *OHLCVAggregator) GetOHLCVData(ctx context.Context, req *GetOhlcvDataReq) (*OHLCV, error) {
	// Validate request
	if req.Limit <= 0 {
		req.Limit = 500
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	// Delegate to repository
	return a.repo.GetOHLCVData(ctx, req)
}

func (a *OHLCVAggregator) GetRealtimeOHLCV(ctx context.Context, symbol, interval string) (*ohlcvBar, error) {
	// First check memory cache
	if symbolBars, ok := a.realtimeBars.Load(symbol); ok {
		if intervalBars, ok := symbolBars.(map[string]*ohlcvBar); ok {
			if bar, exists := intervalBars[interval]; exists {
				return bar, nil
			}
		}
	}

	// Fallback to database
	now := time.Now()
	config := SupportedIntervals[interval]
	openTime := a.getBucketTime(now, config.Duration)

	return a.repo.GetRealtimeOHLCV(ctx, symbol, interval, openTime)
}

// ==================== Health Check ====================

func (a *OHLCVAggregator) GetHealthStatus() map[string]interface{} {
	status := make(map[string]interface{})

	// Count realtime bars
	realtimeCount := 0
	a.realtimeBars.Range(func(k, v interface{}) bool {
		if intervalBars, ok := v.(map[string]*ohlcvBar); ok {
			realtimeCount += len(intervalBars)
		}
		return true
	})

	status["realtime_bars_count"] = realtimeCount
	status["trade_channel_size"] = len(a.tradeCh)
	status["supported_intervals"] = len(SupportedIntervals)
	status["status"] = "running"

	return status
}
