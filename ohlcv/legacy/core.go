package legacy

//import (
//	"context"
//	"database/sql"
//	"fmt"
//	"github.com/labstack/gommon/log"
//	"sync"
//	"time"
//)
//
//// ========================= Config =========================
//
//type IntervalConfig struct {
//	Duration time.Duration
//	Table    string
//}
//
//var SupportedIntervals = map[string]IntervalConfig{
//	"1h": {Duration: time.Hour, Table: "ohlcv_1h"},
//	"1d": {Duration: 24 * time.Hour, Table: "ohlcv_1d"},
//	"1w": {Duration: 7 * 24 * time.Hour, Table: "ohlcv_1w"},
//}
//
//// ========================= New Aggregator =========================
//type AggregatorSetupConfig struct {
//	BatchSize     int
//	FlushInterval time.Duration
//}
//
//type realtimeBars struct {
//	// key: interval, value: bar
//	intervalBars map[string]*OHLCVBar
//	barsMu       sync.RWMutex
//}
//
//func (b *realtimeBars) CloseByOpenTimeAndReNew(openTime int64) []*OHLCVBar {
//	b.barsMu.Lock()
//	defer b.barsMu.Unlock()
//
//	closedBars := make([]*OHLCVBar, 0, len(b.intervalBars))
//
//	for interval, bar := range b.intervalBars {
//		if bar == nil {
//			log.Warnf("nil bar found for interval: %s", interval)
//			continue
//		}
//
//		// Only process bars matching the given openTime
//		if openTime != bar.OpenTime {
//			continue
//		}
//
//		// Mark the bar as closed and collect it
//		bar.IsClosed = true
//		closedBars = append(closedBars, bar)
//
//		// Remove old bar from map
//		delete(b.intervalBars, interval)
//
//		// Retrieve interval configuration
//		cfg, ok := SupportedIntervals[interval]
//		if !ok {
//			log.Warnf("Unsupported interval: %s", interval)
//			continue
//		}
//
//		// Validate duration
//		if cfg.Duration <= 0 {
//			log.Warnf("Invalid duration for interval: %s", interval)
//			continue
//		}
//
//		// Calculate next bar's open and close time
//		nextOpenTime := openTime + int64(cfg.Duration)
//		nextCloseTime := nextOpenTime + int64(cfg.Duration) - 1
//
//		// Create new bar
//		newBar := &OHLCVBar{
//			Symbol:    bar.Symbol,
//			OpenPrice: bar.ClosePrice,
//			HighPrice: bar.ClosePrice,
//			LowPrice:  bar.ClosePrice,
//			Volume:    0.0,
//			OpenTime:  nextOpenTime,
//			CloseTime: nextCloseTime,
//			IsClosed:  false,
//		}
//
//		// Store new bar in the map
//		b.intervalBars[interval] = newBar
//	}
//
//	return closedBars
//}
//
//func newRealtimeBars(intervals map[string]IntervalConfig) *realtimeBars {
//	intervalBars := make(map[string]*OHLCVBar)
//	for interval := range intervals {
//		intervalBars[interval] = &OHLCVBar{}
//	}
//	return &realtimeBars{
//		intervalBars: intervalBars,
//	}
//}
//
//type realtimeBarManager struct {
//	// Real-time OHLCV bars cache (symbol -> interval -> bar)
//	realtimeBars map[string]*realtimeBars
//}
//
//func newRealtimeBarManager() *realtimeBarManager {
//	return &realtimeBarManager{
//		realtimeBars: make(map[string]*realtimeBars),
//	}
//}
//
//func (b *realtimeBarManager) storeBar(symbol string, interval string, bar *OHLCVBar) {
//	if rb, ok := b.realtimeBars[symbol]; ok {
//		rb.barsMu.Lock()
//		defer rb.barsMu.Unlock()
//		rb.intervalBars[interval] = bar
//	} else {
//		// create symbol's bars
//		rb = newRealtimeBars(SupportedIntervals)
//		rb.intervalBars[interval] = bar
//		b.realtimeBars[symbol] = rb
//	}
//}
//
//func (b *realtimeBarManager) Range(f func(symbol string, intervalBarMap map[string]*OHLCVBar) bool) {
//	for symbol, bar := range b.realtimeBars {
//		bar.barsMu.RLock()
//		f(symbol, bar.intervalBars)
//		bar.barsMu.RUnlock()
//	}
//}
//
//func (b *realtimeBarManager) getBarBySymbolAndInterval(symbol string, interval string) (OHLCVBar, bool) {
//	if rbs, ok := b.realtimeBars[symbol]; ok {
//		rbs.barsMu.RLock()
//		defer rbs.barsMu.RUnlock()
//		if bar, ok := rbs.intervalBars[interval]; ok {
//			return *bar, true
//		}
//	}
//	return OHLCVBar{}, false
//}
//
//// closeIntervalBars close expired intervalBar and renew intervalBar, return closed bars
//func (b *realtimeBarManager) closeIntervalBars(ctx context.Context, interval string, openTime int64) []*OHLCVBar {
//	//TODO
//}
//
//type OHLCVAggregator struct {
//	db          *sql.DB
//	repo        OHLCVRepository
//	tradeStream TradeStream
//
//	// Realtime Bars
//	realtimeBarManager *realtimeBarManager
//
//	// Channels for internal communication
//	tradeCh chan *Trade
//	stopCh  chan struct{}
//	// Timers for each interval
//
//	// Configuration
//	batchSize     int
//	flushInterval time.Duration
//	// Statistics tracking
//	statsCache sync.Map // map[string]*OHLCVStatistics
//}
//
//func NewOHLCVAggregator(repo OHLCVRepository, stream TradeStream, config *AggregatorSetupConfig) *OHLCVAggregator {
//	if config.BatchSize <= 0 {
//		// min batch size = 100
//		config.BatchSize = 100
//	}
//	if config.FlushInterval <= 0 {
//		// min flush interval = 5 secs
//		config.FlushInterval = 5 * time.Second
//	}
//
//	return &OHLCVAggregator{
//		repo:               repo,
//		tradeStream:        stream,
//		tradeCh:            make(chan *Trade, 1000),
//		stopCh:             make(chan struct{}),
//		batchSize:          config.BatchSize,
//		flushInterval:      config.FlushInterval,
//		realtimeBarManager: newRealtimeBarManager(),
//	}
//}
//
//// ========================= Aggregator expose func =========================
//
//func (a *OHLCVAggregator) Start(ctx context.Context, symbols []string) error {
//	// Subscribe to trade stream.
//	tradeStreamCh, err := a.tradeStream.Subscribe(ctx, symbols)
//	if err != nil {
//		return fmt.Errorf("[OHLCVAggregator] failed to subscribe trade stream: %w", err)
//	}
//
//	// Start trade processing goroutine
//	go a.processTradeStream(ctx, tradeStreamCh)
//	// Start aggregation goroutine
//	go a.aggregateTrades(ctx)
//	// Start interval timers
//	go a.manageIntervalTimers(ctx)
//	// Start periodic flush
//	go a.periodicFlush(ctx)
//
//	log.Infof("[OHLCVAggregator] OHLCV aggregator started successfully")
//	return nil
//}
//
//func (a *OHLCVAggregator) Stop() error {
//	close(a.stopCh)
//	return a.tradeStream.Close()
//}
//
//// ========================= Aggregator Domain Logic =========================
//
//// processTradeStream receive trade from tradeStreamCh, and push data into internal channel: tradeCh
//func (a *OHLCVAggregator) processTradeStream(ctx context.Context, tradeStreamCh <-chan *Trade) {
//	for {
//		select {
//		case trade := <-tradeStreamCh:
//			if trade != nil {
//				select {
//				case a.tradeCh <- trade:
//				default:
//					log.Warnf("[OHLCVAggregator] Trade channel full, dropping trade: %v", trade)
//				}
//			}
//		case <-ctx.Done():
//			return
//		case <-a.stopCh:
//			return
//		}
//	}
//}
//
//// aggregateTrades listen on tradeCh, receive trade data and process data (batch)
//func (a *OHLCVAggregator) aggregateTrades(ctx context.Context) {
//	// create trade data batch container
//	tradeBatch := make([]*Trade, 0, a.batchSize)
//	ticker := time.NewTicker(a.flushInterval)
//	defer ticker.Stop()
//	for {
//		select {
//		case trade := <-a.tradeCh:
//			tradeBatch = append(tradeBatch, trade)
//
//			// Process batch when full
//			if len(tradeBatch) >= a.batchSize {
//				a.processTradeBatch(ctx, tradeBatch)
//				tradeBatch = tradeBatch[:0] // Reset batch slice
//			}
//
//		case <-ticker.C:
//			// Process remaining trades on timeout
//			if len(tradeBatch) > 0 {
//				a.processTradeBatch(ctx, tradeBatch)
//				tradeBatch = tradeBatch[:0] // Reset batch slice
//			}
//
//		case <-ctx.Done():
//			return
//		case <-a.stopCh:
//			return
//		}
//	}
//}
//
//// processTradeBatch group by symbol and do process.
//func (a *OHLCVAggregator) processTradeBatch(ctx context.Context, trades []*Trade) {
//	// Group trades by symbol
//	symbolTrades := make(map[string][]*Trade)
//	for _, trade := range trades {
//		symbolTrades[trade.Symbol] = append(symbolTrades[trade.Symbol], trade)
//	}
//
//	// Process each symbol's trades
//	for symbol, symbolTradeList := range symbolTrades {
//		a.processSymbolTrades(ctx, symbol, symbolTradeList)
//	}
//}
//
//// processSymbolTrades process trades by (symbol)
//func (a *OHLCVAggregator) processSymbolTrades(ctx context.Context, symbol string, trades []*Trade) {
//	// Process for each supported interval (1h, 1d, 1w...)
//	for interval, config := range SupportedIntervals {
//		a.processTradesForInterval(ctx, symbol, interval, config, trades)
//	}
//}
//
//// processTradesForInterval process trades by (symbol)> (interval)    ex: ETH-UST, 1h
//func (a *OHLCVAggregator) processTradesForInterval(ctx context.Context, symbol, interval string, config IntervalConfig, trades []*Trade) {
//	// Group trades by time buckets
//	buckets := make(map[int64][]*Trade)
//
//	for _, trade := range trades {
//		bucketTime := getBucketTime(trade.Timestamp, config.Duration)
//		buckets[bucketTime] = append(buckets[bucketTime], trade)
//	}
//
//	// Process each bucket
//	for bucketTime, bucketTrades := range buckets {
//		a.updateOHLCVBar(ctx, symbol, interval, bucketTime, config.Duration, bucketTrades)
//	}
//}
//
//// getBucketTime input tradeTime and interval return the timestamp align the interval boundary
//// Example: 1 hr boundary: (1)2024-01-01 00:00:00, (2)2024-01-01 00:01:00, (3)2024-01-01 00:02:00 (4)...
//func getBucketTime(tradeTime time.Time, interval time.Duration) int64 {
//	openTime := tradeTime.Truncate(interval).Unix()
//	return openTime
//}
//
//// getNextBucketTime
//func getNextBucketTime(current time.Time, interval time.Duration) time.Time {
//	next := current.Add(interval)
//	return next.Truncate(interval)
//}
//
//func (a *OHLCVAggregator) updateOHLCVBar(ctx context.Context, symbol, interval string, openTime int64, duration time.Duration, trades []*Trade) {
//	// now we got <openTime> and <closeTime>.
//	closeTime := openTime + int64(duration.Seconds()) - 1
//
//	// Get or create realtime bar by (symbol: ETH-USDT, interval: 1h, openTime: 2024-01-01 14:00:00)
//	var bar *OHLCVBar
//	if existingBar, err := a.repo.GetRealtimeOHLCV(ctx, symbol, interval, openTime); err == nil && existingBar != nil {
//		bar = existingBar
//	} else {
//		// Create new bar with first trade
//		firstTrade := trades[0]
//		bar = &OHLCVBar{
//			Symbol:     symbol,
//			OpenPrice:  firstTrade.Price, // create (o)
//			HighPrice:  firstTrade.Price,
//			LowPrice:   firstTrade.Price,
//			ClosePrice: firstTrade.Price,
//			Volume:     0,
//			OpenTime:   openTime,
//			CloseTime:  closeTime,
//			IsClosed:   false,
//		}
//	}
//
//	// Update bar with all trades
//	for _, trade := range trades {
//		// update h, l, c, v
//		a.updateBarWithTrade(bar, trade)
//	}
//
//	// Save/update realtime bar
//	if err := a.repo.UpdateRealtimeOHLCV(ctx, bar, interval); err != nil {
//		log.Errorf("[OHLCVAggregator] Failed to update realtime OHLCV: %v", err)
//	}
//
//	// Store in memory cache for quick access
//	//a.storeRealtimeBar(symbol, interval, bar)
//	a.realtimeBarManager.storeBar(symbol, interval, bar)
//}
//
//// updateBarWithTrade update h, l, c, v
//func (a *OHLCVAggregator) updateBarWithTrade(bar *OHLCVBar, trade *Trade) {
//	// Update high (h)
//	bar.HighPrice = max(bar.HighPrice, trade.Price)
//	// update low (l)
//	bar.LowPrice = min(bar.LowPrice, trade.Price)
//	// update close (c)
//	bar.ClosePrice = trade.Price
//	// Update volume (v)
//	bar.Volume += trade.Volume
//	bar.QuoteVolume += trade.Volume * trade.Price
//	bar.TradeCount++
//}
//
//// ==================== Interval Timer Management ====================
//
//func (a *OHLCVAggregator) manageIntervalTimers(ctx context.Context) {
//	for interval, config := range SupportedIntervals {
//		go a.startIntervalTimer(ctx, interval, config)
//	}
//}
//
//// startIntervalTimer process interval (1h, 1d, 1w, ...), if reached close bar time, do closeIntervalBars()
//func (a *OHLCVAggregator) startIntervalTimer(ctx context.Context, interval string, config IntervalConfig) {
//	// Calculate next interval boundary
//	now := time.Now()
//	nextBoundary := getNextBucketTime(now, config.Duration)
//
//	timer := time.NewTimer(time.Until(nextBoundary))
//	defer timer.Stop()
//
//	for {
//		select {
//		case <-timer.C:
//			a.closeIntervalBars(ctx, interval, nextBoundary.Add(-config.Duration).Unix())
//			// Set next timer
//			nextBoundary = nextBoundary.Add(config.Duration)
//			timer.Reset(time.Until(nextBoundary))
//
//		case <-ctx.Done():
//			return
//		case <-a.stopCh:
//			return
//		}
//	}
//}
//
//func (a *OHLCVAggregator) closeIntervalBars(ctx context.Context, interval string, openTime int64) {
//	closedBars := a.realtimeBarManager.closeIntervalBars(ctx, interval, openTime)
//	if err := a.repo.SaveOHLCVBars(ctx, closedBars, interval); err != nil {
//		log.Errorf("[OHLCVAggregator] Failed to save OHLCVBars: %v", err)
//	}
//}
//
//// ==================== Periodic Tasks ====================
//
//func (a *OHLCVAggregator) periodicFlush(ctx context.Context) {
//	ticker := time.NewTicker(time.Minute * 5) // Flush every 5 minutes
//	defer ticker.Stop()
//
//	for {
//		select {
//		case <-ticker.C:
//			a.flushRealtimeBars(ctx)
//		case <-ctx.Done():
//			return
//		case <-a.stopCh:
//			return
//		}
//	}
//}
//
//func (a *OHLCVAggregator) flushRealtimeBars(ctx context.Context) {
//	a.realtimeBarManager.Range(func(symbol string, intervalBarMap map[string]*OHLCVBar) bool {
//		for interval, bar := range intervalBarMap {
//			if err := a.repo.UpdateRealtimeOHLCV(ctx, bar, interval); err != nil {
//				log.Warnf("[OHLCVAggregator] Failed to flush realtime bar: %v", err)
//			}
//		}
//		return true
//	})
//}
//
//// ==================== Public Query Methods ====================
//
//func (a *OHLCVAggregator) GetOHLCVData(ctx context.Context, req *GetOhlcvDataReq) (*OHLCV, error) {
//	// Validate request
//	if req.Limit <= 0 {
//		req.Limit = 500
//	}
//	if req.Limit > 1000 {
//		req.Limit = 1000
//	}
//
//	// Delegate to repository
//	return a.repo.GetOHLCVData(ctx, req)
//}
//
//func (a *OHLCVAggregator) GetRealtimeOHLCV(ctx context.Context, symbol, interval string) (OHLCVBar, error) {
//	// First check memory cache
//	if ohlcvBar, ok := a.realtimeBarManager.getBarBySymbolAndInterval(symbol, interval); ok {
//		return ohlcvBar, nil
//	}
//
//	// Fallback to database
//	now := time.Now()
//	config := SupportedIntervals[interval]
//	openTime := getBucketTime(now, config.Duration)
//
//	ohlcvBar, err := a.repo.GetRealtimeOHLCV(ctx, symbol, interval, openTime)
//
//	return *ohlcvBar, err
//}
//
//// ==================== Health Check ====================
//
//func (a *OHLCVAggregator) GetHealthStatus() map[string]interface{} {
//	status := make(map[string]interface{})
//
//	// Count realtime bars
//	realtimeCount := 0
//	a.realtimeBarManager.Range(func(symbol string, intervalBars map[string]*OHLCVBar) bool {
//		realtimeCount += len(intervalBars)
//		return true
//	})
//
//	status["realtime_bars_count"] = realtimeCount
//	status["trade_channel_size"] = len(a.tradeCh)
//	status["supported_intervals"] = len(SupportedIntervals)
//	status["status"] = "running"
//
//	return status
//}
