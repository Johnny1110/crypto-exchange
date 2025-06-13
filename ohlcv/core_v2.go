package ohlcv

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/labstack/gommon/log"
)

// ========================= Constants & Configuration =========================

const (
	DefaultBatchSize     = 100
	MinBatchSize         = 10
	MaxBatchSize         = 1000
	DefaultFlushInterval = 5 * time.Second
	MinFlushInterval     = time.Second
	MaxFlushInterval     = time.Minute
	DefaultChannelSize   = 1000
	MaxChannelSize       = 10000
)

type IntervalConfig struct {
	Duration time.Duration
	Table    string
}

var SupportedIntervals = map[string]IntervalConfig{
	"1h": {Duration: time.Hour, Table: "ohlcv_1h"},
	"1d": {Duration: 24 * time.Hour, Table: "ohlcv_1d"},
	"1w": {Duration: 7 * 24 * time.Hour, Table: "ohlcv_1w"},
	// add more if needed.
}

// ========================= Configuration & Validation =========================

type Config struct {
	BatchSize      int
	FlushInterval  time.Duration
	ChannelSize    int
	MaxConcurrency int
	EnableMetrics  bool
}

func (c *Config) Validate() error {
	if c.BatchSize < MinBatchSize || c.BatchSize > MaxBatchSize {
		return fmt.Errorf("invalid batch size: %d, must be between %d and %d",
			c.BatchSize, MinBatchSize, MaxBatchSize)
	}
	if c.FlushInterval < MinFlushInterval || c.FlushInterval > MaxFlushInterval {
		return fmt.Errorf("invalid flush interval: %v, must be between %v and %v",
			c.FlushInterval, MinFlushInterval, MaxFlushInterval)
	}
	if c.ChannelSize <= 0 || c.ChannelSize > MaxChannelSize {
		return fmt.Errorf("invalid channel size: %d, must be between 1 and %d",
			c.ChannelSize, MaxChannelSize)
	}
	if c.MaxConcurrency <= 0 {
		c.MaxConcurrency = 10
	}
	return nil
}

func DefaultConfig() *Config {
	return &Config{
		BatchSize:      DefaultBatchSize,
		FlushInterval:  DefaultFlushInterval,
		ChannelSize:    DefaultChannelSize,
		MaxConcurrency: 10,
		EnableMetrics:  true,
	}
}

// ========================= Metrics & Monitoring =========================

type Metrics struct {
	tradesProcessed    int64
	barsCreated        int64
	barsUpdated        int64
	barsClosed         int64
	errorsCount        int64
	lastProcessTime    int64
	channelUtilization int64
}

func (m *Metrics) IncrementTradesProcessed() {
	atomic.AddInt64(&m.tradesProcessed, 1)
}

func (m *Metrics) IncrementBarsCreated() {
	atomic.AddInt64(&m.barsCreated, 1)
}

func (m *Metrics) IncrementBarsUpdated() {
	atomic.AddInt64(&m.barsUpdated, 1)
}

func (m *Metrics) IncrementBarsClosed() {
	atomic.AddInt64(&m.barsClosed, 1)
}

func (m *Metrics) IncrementErrors() {
	atomic.AddInt64(&m.errorsCount, 1)
}

func (m *Metrics) UpdateProcessTime() {
	atomic.StoreInt64(&m.lastProcessTime, time.Now().Unix())
}

func (m *Metrics) UpdateChannelUtilization(current, capacity int) {
	utilization := int64((current * 100) / capacity)
	atomic.StoreInt64(&m.channelUtilization, utilization)
}

func (m *Metrics) GetSnapshot() map[string]int64 {
	return map[string]int64{
		"trades_processed":    atomic.LoadInt64(&m.tradesProcessed),
		"bars_created":        atomic.LoadInt64(&m.barsCreated),
		"bars_updated":        atomic.LoadInt64(&m.barsUpdated),
		"bars_closed":         atomic.LoadInt64(&m.barsClosed),
		"errors_count":        atomic.LoadInt64(&m.errorsCount),
		"last_process_time":   atomic.LoadInt64(&m.lastProcessTime),
		"channel_utilization": atomic.LoadInt64(&m.channelUtilization),
	}
}

// ========================= Safe Bar Management =========================

type SafeBar struct {
	bar   *OHLCVBar
	mutex sync.RWMutex
	dirty bool
}

func NewSafeBar(bar *OHLCVBar) *SafeBar {
	return &SafeBar{
		bar:   bar,
		dirty: false,
	}
}

func (sb *SafeBar) Update(trade *Trade) {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	if sb.bar.OpenPrice == 0 {
		sb.bar.OpenPrice = trade.Price
	}

	if trade.Price > sb.bar.HighPrice {
		sb.bar.HighPrice = trade.Price
	}
	if trade.Price < sb.bar.LowPrice || sb.bar.LowPrice == 0 {
		sb.bar.LowPrice = trade.Price
	}

	sb.bar.ClosePrice = trade.Price
	sb.bar.Volume += trade.Volume
	sb.bar.QuoteVolume += trade.Volume * trade.Price
	sb.bar.TradeCount++
	sb.dirty = true
}

func (sb *SafeBar) GetCopy() *OHLCVBar {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()

	barCopy := *sb.bar
	return &barCopy
}

func (sb *SafeBar) MarkClean() {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()
	sb.dirty = false
}

func (sb *SafeBar) IsDirty() bool {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()
	return sb.dirty
}

func (sb *SafeBar) Close() *OHLCVBar {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	sb.bar.IsClosed = true
	closedBar := *sb.bar
	return &closedBar
}

// ========================= Interval Bar Manager =========================

type IntervalBars struct {
	bars   map[string]*SafeBar // interval -> SafeBar
	symbol string
	mutex  sync.RWMutex
}

func NewIntervalBars(symbol string) *IntervalBars {
	return &IntervalBars{
		bars:   make(map[string]*SafeBar),
		symbol: symbol,
	}
}

func (ib *IntervalBars) GetOrCreate(interval string, openTime, closeTime int64, firstPrice float64) *SafeBar {
	ib.mutex.Lock()
	defer ib.mutex.Unlock()

	bar, exists := ib.bars[interval]
	if !exists {
		ohlcvBar := &OHLCVBar{
			Symbol:     ib.symbol,
			OpenPrice:  firstPrice,
			HighPrice:  firstPrice,
			LowPrice:   firstPrice,
			ClosePrice: firstPrice,
			Volume:     0,
			OpenTime:   openTime,
			CloseTime:  closeTime,
			IsClosed:   false,
		}
		bar = NewSafeBar(ohlcvBar)
		ib.bars[interval] = bar
	}
	return bar
}

func (ib *IntervalBars) Get(interval string) (*SafeBar, bool) {
	ib.mutex.RLock()
	defer ib.mutex.RUnlock()

	bar, exists := ib.bars[interval]
	return bar, exists
}

func (ib *IntervalBars) CloseAndRenew(interval string, openTime int64) *OHLCVBar {
	ib.mutex.Lock()
	defer ib.mutex.Unlock()

	bar, exists := ib.bars[interval]
	if !exists || bar.bar.OpenTime != openTime {
		return nil
	}

	closedBar := bar.Close()
	delete(ib.bars, interval)

	// Create new bar for next interval
	config, ok := SupportedIntervals[interval]
	if ok {
		nextOpenTime := openTime + int64(config.Duration.Seconds())
		nextCloseTime := nextOpenTime + int64(config.Duration.Seconds()) - 1

		newOHLCVBar := &OHLCVBar{
			Symbol:     ib.symbol,
			OpenPrice:  closedBar.ClosePrice,
			HighPrice:  closedBar.ClosePrice,
			LowPrice:   closedBar.ClosePrice,
			ClosePrice: closedBar.ClosePrice,
			Volume:     0,
			OpenTime:   nextOpenTime,
			CloseTime:  nextCloseTime,
			IsClosed:   false,
		}
		ib.bars[interval] = NewSafeBar(newOHLCVBar)
	}

	return closedBar
}

func (ib *IntervalBars) GetAllDirtyBars() []*SafeBar {
	ib.mutex.RLock()
	defer ib.mutex.RUnlock()

	var dirtyBars []*SafeBar
	for _, bar := range ib.bars {
		if bar.IsDirty() {
			dirtyBars = append(dirtyBars, bar)
		}
	}
	return dirtyBars
}

// ========================= Bar Manager =========================

type BarManager struct {
	symbolBars map[string]*IntervalBars // symbol -> IntervalBars
	mutex      sync.RWMutex
}

func NewBarManager() *BarManager {
	return &BarManager{
		symbolBars: make(map[string]*IntervalBars),
	}
}

func (bm *BarManager) GetOrCreateIntervalBars(symbol string) *IntervalBars {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	ib, exists := bm.symbolBars[symbol]
	if !exists {
		ib = NewIntervalBars(symbol)
		bm.symbolBars[symbol] = ib
	}
	return ib
}

func (bm *BarManager) CloseExpiredBars(interval string, openTime int64) []*OHLCVBar {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	var closedBars []*OHLCVBar
	for _, intervalBars := range bm.symbolBars {
		if closedBar := intervalBars.CloseAndRenew(interval, openTime); closedBar != nil {
			closedBars = append(closedBars, closedBar)
		}
	}
	return closedBars
}

func (bm *BarManager) GetRealtimeBar(symbol, interval string) (*OHLCVBar, bool) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	intervalBars, exists := bm.symbolBars[symbol]
	if !exists {
		return nil, false
	}

	safeBar, exists := intervalBars.Get(interval)
	if !exists {
		return nil, false
	}

	return safeBar.GetCopy(), true
}

func (bm *BarManager) FlushDirtyBars() map[string][]*SafeBar {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	dirtyBars := make(map[string][]*SafeBar)
	for symbol, intervalBars := range bm.symbolBars {
		bars := intervalBars.GetAllDirtyBars()
		if len(bars) > 0 {
			dirtyBars[symbol] = bars
		}
	}
	return dirtyBars
}

func (bm *BarManager) GetHealthStatus() map[string]interface{} {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	status := map[string]interface{}{
		"symbols_count": len(bm.symbolBars),
		"total_bars":    0,
	}

	totalBars := 0
	for _, intervalBars := range bm.symbolBars {
		intervalBars.mutex.RLock()
		totalBars += len(intervalBars.bars)
		intervalBars.mutex.RUnlock()
	}
	status["total_bars"] = totalBars

	return status
}

// ========================= Worker Pool =========================

type WorkerPool struct {
	workers   int
	taskCh    chan func()
	stopCh    chan struct{}
	wg        sync.WaitGroup
	isRunning int32
}

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		taskCh:  make(chan func(), workers*2),
		stopCh:  make(chan struct{}),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&wp.isRunning, 0, 1) {
		return
	}

	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx)
	}
}

func (wp *WorkerPool) worker(ctx context.Context) {
	defer wp.wg.Done()

	for {
		select {
		case task := <-wp.taskCh:
			if task != nil {
				task()
			}
		case <-ctx.Done():
			return
		case <-wp.stopCh:
			return
		}
	}
}

func (wp *WorkerPool) Submit(task func()) bool {
	if atomic.LoadInt32(&wp.isRunning) == 0 {
		return false
	}

	select {
	case wp.taskCh <- task:
		return true
	default:
		return false
	}
}

func (wp *WorkerPool) Stop() {
	if !atomic.CompareAndSwapInt32(&wp.isRunning, 1, 0) {
		return
	}

	close(wp.stopCh)
	wp.wg.Wait()
}

// ========================= Main Aggregator =========================

type OHLCVAggregator struct {
	// Dependencies
	repo        OHLCVRepository
	tradeStream TradeStream

	// Core components
	barManager *BarManager
	workerPool *WorkerPool
	metrics    *Metrics

	// Configuration
	config *Config

	// Channels
	tradeCh chan *Trade
	stopCh  chan struct{}

	// State management
	isRunning int32

	// Timers
	intervalTimers map[string]*time.Timer
	timerMutex     sync.RWMutex
}

func NewOHLCVAggregator(repo OHLCVRepository, stream TradeStream, config *Config) (*OHLCVAggregator, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &OHLCVAggregator{
		repo:           repo,
		tradeStream:    stream,
		barManager:     NewBarManager(),
		workerPool:     NewWorkerPool(config.MaxConcurrency),
		metrics:        &Metrics{},
		config:         config,
		tradeCh:        make(chan *Trade, config.ChannelSize),
		stopCh:         make(chan struct{}),
		intervalTimers: make(map[string]*time.Timer),
	}, nil
}

// ========================= Public Methods =========================

func (a *OHLCVAggregator) Start(ctx context.Context, symbols []string) error {
	if !atomic.CompareAndSwapInt32(&a.isRunning, 0, 1) {
		return fmt.Errorf("aggregator is already running")
	}

	if len(symbols) == 0 {
		return fmt.Errorf("no symbols provided")
	}

	// Start worker pool
	a.workerPool.Start(ctx)

	// Subscribe to trade stream
	tradeStreamCh, err := a.tradeStream.Subscribe(ctx, symbols)
	if err != nil {
		atomic.StoreInt32(&a.isRunning, 0)
		return fmt.Errorf("failed to subscribe trade stream: %w", err)
	}

	// Start core goroutines
	go a.processTradeStream(ctx, tradeStreamCh)
	go a.aggregateTrades(ctx)
	go a.manageIntervalTimers(ctx)
	go a.periodicFlush(ctx)

	if a.config.EnableMetrics {
		go a.metricsCollector(ctx)
	}

	log.Infof("OHLCV aggregator started successfully with %d symbols", len(symbols))
	return nil
}

func (a *OHLCVAggregator) Stop() error {
	if !atomic.CompareAndSwapInt32(&a.isRunning, 1, 0) {
		return fmt.Errorf("aggregator is not running")
	}

	close(a.stopCh)
	a.workerPool.Stop()

	// Stop interval timers
	a.timerMutex.Lock()
	for _, timer := range a.intervalTimers {
		timer.Stop()
	}
	a.timerMutex.Unlock()

	return a.tradeStream.Close()
}

func (a *OHLCVAggregator) IsRunning() bool {
	return atomic.LoadInt32(&a.isRunning) == 1
}

// ========================= Core Processing Logic =========================

func (a *OHLCVAggregator) processTradeStream(ctx context.Context, tradeStreamCh <-chan *Trade) {
	defer log.Info("Trade stream processor stopped")

	for {
		select {
		case trade := <-tradeStreamCh:
			if trade == nil {
				continue
			}

			// Validate trade
			if err := a.validateTrade(trade); err != nil {
				log.Warnf("Invalid trade received: %v, error: %v", trade, err)
				a.metrics.IncrementErrors()
				continue
			}

			select {
			case a.tradeCh <- trade:
				// Update channel utilization metric
				if a.config.EnableMetrics {
					a.metrics.UpdateChannelUtilization(len(a.tradeCh), cap(a.tradeCh))
				}
			default:
				log.Warnf("Trade channel full, dropping trade: %v", trade)
				a.metrics.IncrementErrors()
			}

		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		}
	}
}

func (a *OHLCVAggregator) aggregateTrades(ctx context.Context) {
	defer log.Info("Trade aggregator stopped")

	tradeBatch := make([]*Trade, 0, a.config.BatchSize)
	ticker := time.NewTicker(a.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case trade := <-a.tradeCh:
			tradeBatch = append(tradeBatch, trade)

			if len(tradeBatch) >= a.config.BatchSize {
				a.processTradeBatch(ctx, tradeBatch)
				tradeBatch = tradeBatch[:0]
			}

		case <-ticker.C:
			if len(tradeBatch) > 0 {
				a.processTradeBatch(ctx, tradeBatch)
				tradeBatch = tradeBatch[:0]
			}

		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		}
	}
}

func (a *OHLCVAggregator) processTradeBatch(ctx context.Context, trades []*Trade) {
	startTime := time.Now()
	defer func() {
		a.metrics.UpdateProcessTime()
		log.Debugf("Processed %d trades in %v", len(trades), time.Since(startTime))
	}()

	// Group trades by symbol for better processing
	symbolTrades := make(map[string][]*Trade)
	for _, trade := range trades {
		symbolTrades[trade.Symbol] = append(symbolTrades[trade.Symbol], trade)
		a.metrics.IncrementTradesProcessed()
	}

	// Process each symbol's trades concurrently
	for symbol, trades := range symbolTrades {
		symbol, trades := symbol, trades // capture loop variables

		task := func() {
			a.processSymbolTrades(ctx, symbol, trades)
		}

		if !a.workerPool.Submit(task) {
			// Fallback to synchronous processing if worker pool is full
			log.Warn("Worker pool full, processing synchronously")
			task()
		}
	}
}

func (a *OHLCVAggregator) processSymbolTrades(ctx context.Context, symbol string, trades []*Trade) {
	intervalBars := a.barManager.GetOrCreateIntervalBars(symbol)

	for interval, config := range SupportedIntervals {
		a.processTradesForInterval(ctx, symbol, interval, config, trades, intervalBars)
	}
}

func (a *OHLCVAggregator) processTradesForInterval(ctx context.Context, symbol, interval string, config IntervalConfig, trades []*Trade, intervalBars *IntervalBars) {
	// Group trades by time buckets
	buckets := make(map[int64][]*Trade)

	for _, trade := range trades {
		bucketTime := getBucketTime(trade.Timestamp, config.Duration)
		buckets[bucketTime] = append(buckets[bucketTime], trade)
	}

	// Process each bucket
	for bucketTime, bucketTrades := range buckets {
		a.updateOHLCVBar(ctx, symbol, interval, bucketTime, config.Duration, bucketTrades, intervalBars)
	}
}

func (a *OHLCVAggregator) updateOHLCVBar(ctx context.Context, symbol, interval string, openTime int64, duration time.Duration, trades []*Trade, intervalBars *IntervalBars) {
	closeTime := openTime + int64(duration.Seconds()) - 1
	firstTrade := trades[0]

	// Get or create safe bar
	safeBar := intervalBars.GetOrCreate(interval, openTime, closeTime, firstTrade.Price)

	// Update bar with all trades
	for _, trade := range trades {
		safeBar.Update(trade)
	}

	a.metrics.IncrementBarsUpdated()
}

// ========================= Timer Management =========================

func (a *OHLCVAggregator) manageIntervalTimers(ctx context.Context) {
	defer log.Info("Interval timer manager stopped")

	for interval, config := range SupportedIntervals {
		interval, config := interval, config // capture loop variables
		go a.startIntervalTimer(ctx, interval, config)
	}

	<-ctx.Done()
}

func (a *OHLCVAggregator) startIntervalTimer(ctx context.Context, interval string, config IntervalConfig) {
	now := time.Now()
	nextBoundary := getNextBucketTime(now, config.Duration)

	timer := time.NewTimer(time.Until(nextBoundary))

	a.timerMutex.Lock()
	a.intervalTimers[interval] = timer
	a.timerMutex.Unlock()

	defer func() {
		timer.Stop()
		a.timerMutex.Lock()
		delete(a.intervalTimers, interval)
		a.timerMutex.Unlock()
	}()

	for {
		select {
		case <-timer.C:
			openTime := nextBoundary.Add(-config.Duration).Unix()
			a.closeIntervalBars(ctx, interval, openTime)

			nextBoundary = nextBoundary.Add(config.Duration)
			timer.Reset(time.Until(nextBoundary))

		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		}
	}
}

func (a *OHLCVAggregator) closeIntervalBars(ctx context.Context, interval string, openTime int64) {
	closedBars := a.barManager.CloseExpiredBars(interval, openTime)

	if len(closedBars) > 0 {
		if err := a.repo.SaveOHLCVBars(ctx, closedBars, interval); err != nil {
			log.Errorf("Failed to save OHLCV bars: %v", err)
			a.metrics.IncrementErrors()
		} else {
			a.metrics.IncrementBarsClosed()
			log.Debugf("Closed and saved %d bars for interval %s", len(closedBars), interval)
		}
	}
}

// ========================= Periodic Tasks =========================

func (a *OHLCVAggregator) periodicFlush(ctx context.Context) {
	defer log.Info("Periodic flush stopped")

	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.flushRealtimeBars(ctx)
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		}
	}
}

func (a *OHLCVAggregator) flushRealtimeBars(ctx context.Context) {
	dirtyBars := a.barManager.FlushDirtyBars()

	for symbol, bars := range dirtyBars {
		for _, safeBar := range bars {
			bar := safeBar.GetCopy()
			if err := a.repo.UpdateRealtimeOHLCV(ctx, bar, symbol); err != nil {
				log.Warnf("Failed to flush realtime bar for %s: %v", symbol, err)
				a.metrics.IncrementErrors()
			} else {
				safeBar.MarkClean()
			}
		}
	}
}

func (a *OHLCVAggregator) metricsCollector(ctx context.Context) {
	defer log.Info("Metrics collector stopped")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := a.metrics.GetSnapshot()
			log.Infof("Aggregator metrics: %+v", metrics)

		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		}
	}
}

// ========================= Utility Functions =========================

func (a *OHLCVAggregator) validateTrade(trade *Trade) error {
	if trade.Symbol == "" {
		return fmt.Errorf("empty symbol")
	}
	if trade.Price <= 0 {
		return fmt.Errorf("invalid price: %f", trade.Price)
	}
	if trade.Volume <= 0 {
		return fmt.Errorf("invalid volume: %f", trade.Volume)
	}
	if trade.Timestamp.IsZero() {
		return fmt.Errorf("invalid timestamp")
	}
	return nil
}

func getBucketTime(tradeTime time.Time, interval time.Duration) int64 {
	return tradeTime.Truncate(interval).Unix()
}

func getNextBucketTime(current time.Time, interval time.Duration) time.Time {
	return current.Add(interval).Truncate(interval)
}

// ========================= Public Query Methods =========================

func (a *OHLCVAggregator) GetOHLCVData(ctx context.Context, req *GetOhlcvDataReq) (*OHLCV, error) {
	if req.Limit <= 0 {
		req.Limit = 500
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	return a.repo.GetOHLCVData(ctx, req)
}

func (a *OHLCVAggregator) GetRealtimeOHLCV(ctx context.Context, symbol, interval string) (*OHLCVBar, error) {
	// Check memory cache first
	if bar, exists := a.barManager.GetRealtimeBar(symbol, interval); exists {
		return bar, nil
	}

	// Fallback to database
	now := time.Now()
	config, exists := SupportedIntervals[interval]
	if !exists {
		return nil, fmt.Errorf("unsupported interval: %s", interval)
	}

	openTime := getBucketTime(now, config.Duration)
	return a.repo.GetRealtimeOHLCV(ctx, symbol, interval, openTime)
}

// ==================== Health Check ====================
