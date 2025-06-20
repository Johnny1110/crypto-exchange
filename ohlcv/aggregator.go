package ohlcv

import (
	"context"
	"fmt"
	"github.com/labstack/gommon/log"
	"sync"
	"time"
)

type OHLCVAggregator struct {
	// Dependencies
	repo        OHLCVRepository
	tradeStream TradeStream

	// Core components
	realtimeSymbolBars sync.Map
	workerPool         *WorkerPool

	// Config
	config *AggregatorConfig

	// Channels
	tradeCh chan *Trade
	stopCh  chan struct{}

	// State management
	isRunning int32
}

func NewOHLCVAggregator(repo OHLCVRepository, stream TradeStream, config *AggregatorConfig) (*OHLCVAggregator, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository cannot be nil")
	}

	if stream == nil {
		return nil, fmt.Errorf("trade stream cannot be nil")
	}

	if config == nil {
		config = DefaultAggregatorConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	wp := NewWorkerPool(config.MaxConcurrency)
	wp.Start(context.Background())

	return &OHLCVAggregator{
		repo:               repo,
		tradeStream:        stream,
		realtimeSymbolBars: sync.Map{},
		workerPool:         wp,
		config:             config,
		tradeCh:            make(chan *Trade, config.ChannelSize),
		stopCh:             make(chan struct{}),
		isRunning:          0,
	}, nil
}

// AddSymbol defaultConfigs could be nil (using default)
func (agg *OHLCVAggregator) AddSymbol(symbol string, initPrice float64, defaultConfigs map[OHLCV_INTERVAL]IntervalConfig) error {
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	if initPrice < 0 {
		return fmt.Errorf("init price must be positive")
	}

	if _, exists := agg.realtimeSymbolBars.Load(symbol); exists {
		return fmt.Errorf("symbol %s already exists", symbol)
	}

	if defaultConfigs == nil {
		defaultConfigs = SupportedIntervals
	}

	// new RealtimeSymbolBars
	symbolBars := NewRealtimeSymbolBars(symbol, initPrice, defaultConfigs)
	agg.realtimeSymbolBars.Store(symbol, symbolBars)

	return nil
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
	// Start interval timers
	go a.manageIntervalTimers(ctx)
	// Start periodic flush
	go a.periodicFlush(ctx)

	a.isRunning = 1
	log.Infof("[OHLCVAggregator] OHLCV aggregator started successfully")
	return nil
}

func (a *OHLCVAggregator) Stop() error {
	close(a.stopCh)
	a.workerPool.Stop()
	return a.tradeStream.Close()
}

func (a *OHLCVAggregator) processTradeStream(ctx context.Context, ch <-chan *Trade) {
	for {
		select {
		case trade := <-ch:
			if trade != nil {
				select {
				case a.tradeCh <- trade:
				default:
					log.Warnf("[OHLCVAggregator] Trade channel full, dropping trade: %v", trade)
				}
			}
		case <-ctx.Done():
			log.Infof("[OHLCVAggregator] OHLCV aggregator processTradeStream stopped by context done.")
			return
		case <-a.stopCh:
			log.Infof("[OHLCVAggregator] OHLCV aggregator processTradeStream stopped by stop channel.")
			return
		}
	}
}

func (a *OHLCVAggregator) aggregateTrades(ctx context.Context) {
	// create trade data batch container
	tradeBatch := make([]*Trade, 0, a.config.BatchSize)
	log.Infof("[OHLCVAggregator] flash batch trades interval: %v", a.config.FlushInterval)
	ticker := time.NewTicker(a.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case trade := <-a.tradeCh:
			log.Debugf("[OHLCVAggregator] recieved trade %v, put into tradeBatch.", trade)
			tradeBatch = append(tradeBatch, trade)
			if len(tradeBatch) >= a.config.BatchSize {
				a.processTradeBatch(ctx, tradeBatch)
				tradeBatch = tradeBatch[:0]
			}
		case <-ticker.C:
			log.Debugf("[OHLCVAggregator] try to flush trade batch, data count: %v", len(tradeBatch))
			if len(tradeBatch) > 0 {
				a.processTradeBatch(ctx, tradeBatch)
				tradeBatch = tradeBatch[:0]
			}
		case <-ctx.Done():
			log.Infof("[OHLCVAggregator] aggregator aggregateTrades stopped by context done.")
			return
		case <-a.stopCh:
			log.Infof("[OHLCVAggregator] aggregator aggregateTrades stopped by stop channel.")
			return
		}
	}
}

// ========================= Aggregator main logic =========================

func (a *OHLCVAggregator) processTradeBatch(ctx context.Context, trades []*Trade) {
	// Group trades by symbol
	symbolTrades := make(map[string][]*Trade)
	for _, trade := range trades {
		symbolTrades[trade.Symbol] = append(symbolTrades[trade.Symbol], trade)
	}

	// Each symbol's trades can be processed concurrently
	for symbol, ts := range symbolTrades {
		if value, ok := a.realtimeSymbolBars.Load(symbol); ok {
			symbolBars := value.(*RealtimeSymbolBars)
			a.workerPool.Submit(func() {
				symbolBars.UpdateByTrades(ctx, ts)
			})
		}
	}
}

// ==================== Interval Timer Management ====================

func (a *OHLCVAggregator) manageIntervalTimers(ctx context.Context) {
	for interval, config := range SupportedIntervals {
		go a.startIntervalTimer(ctx, interval, config)
	}
}

// startIntervalTimer process interval (1h, 1d, 1w, ...), if reached close bar time, do closeIntervalBars()
func (a *OHLCVAggregator) startIntervalTimer(ctx context.Context, interval OHLCV_INTERVAL, config IntervalConfig) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	var lastClosedTimestamp int64

	for {
		select {
		case now := <-ticker.C:
			// check if there has bucket to close.
			bucketToClose := a.findBucketToClose(now, config.Duration, lastClosedTimestamp)

			if bucketToClose > 0 && bucketToClose > lastClosedTimestamp {
				log.Infof("[OHLCVAggregator] Closing interval bar for timestamp %d (%s)",
					bucketToClose,
					time.Unix(bucketToClose, 0).Format("2006-01-02 15:04:05"))

				a.closeIntervalBars(ctx, interval, bucketToClose)
				lastClosedTimestamp = bucketToClose
			}

		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		}
	}
}

func (a *OHLCVAggregator) findBucketToClose(now time.Time, interval time.Duration, lastClosedTimestamp int64) int64 {
	currentBucket := getBucketTime(now, interval)

	previousBucket := currentBucket.Add(-interval)
	previousBucketEnd := previousBucket.Add(interval - 1*time.Second)

	if (now.After(previousBucketEnd) || now.Equal(previousBucketEnd)) &&
		previousBucket.Unix() > lastClosedTimestamp {
		return previousBucket.Unix()
	}

	return 0
}

func (a *OHLCVAggregator) closeIntervalBars(ctx context.Context, interval OHLCV_INTERVAL, openTime int64) {
	a.realtimeSymbolBars.Range(func(key, value interface{}) bool {
		symbol := key.(string)
		rsBars := value.(*RealtimeSymbolBars)
		closedBars, err := rsBars.CloseBars(interval, openTime)
		log.Infof("[OHLCVAggregator] attampt to close OHLCV bar, symbol:%s, interval:%v count:%v", symbol, interval, len(closedBars))
		if err = a.repo.UpsertOHLCVBars(ctx, closedBars, interval); err != nil {
			log.Errorf("[OHLCVAggregator] Failed to save OHLCVBars: %v", err)
		}
		return true
	})
}

// ==================== Periodic Tasks ====================

func (a *OHLCVAggregator) periodicFlush(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 1) // Flush every 1 minutes
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
	a.realtimeSymbolBars.Range(func(key, value interface{}) bool {
		symbolBars := value.(*RealtimeSymbolBars)

		for _, interval := range symbolBars.GetAllIntervals() {
			if bar, ok := symbolBars.GetIntervalBar(interval); ok {
				if err := a.repo.UpdateRealtimeOHLCV(ctx, bar, interval); err != nil {
					log.Warnf("[OHLCVAggregator] Failed to flush realtime bar: %v", err)
				}
			}
		}

		return true
	})
}

// ======================================== Public Query Methods ========================================

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

func (a *OHLCVAggregator) GetRealtimeOHLCV(ctx context.Context, symbol string, interval OHLCV_INTERVAL) (OHLCVBar, error) {
	// First check memory cache
	if ohlcvBar, ok := a.realtimeSymbolBars.Load(symbol); ok {
		rtsBars := ohlcvBar.(*RealtimeSymbolBars)
		if bar, ok := rtsBars.GetIntervalBar(interval); ok {
			return bar, nil
		} else {
			return OHLCVBar{}, fmt.Errorf("failed to get realtime OHLCV bar for symbol: %v, interval: %v", symbol, interval)
		}
	} else {
		return OHLCVBar{}, fmt.Errorf("realtime OHLCV bar not found for symbol: %v", symbol)
	}
}
