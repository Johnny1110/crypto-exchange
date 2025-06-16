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

	// Timers
	intervalTimers map[string]*time.Timer
	timerMutex     sync.RWMutex
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

	return &OHLCVAggregator{
		repo:               repo,
		tradeStream:        stream,
		realtimeSymbolBars: sync.Map{},
		workerPool:         NewWorkerPool(config.MaxConcurrency),
		config:             config,
		tradeCh:            make(chan *Trade, config.ChannelSize),
		stopCh:             make(chan struct{}),
		intervalTimers:     make(map[string]*time.Timer),
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

	log.Infof("[OHLCVAggregator] OHLCV aggregator started successfully")
	return nil
}

func (a *OHLCVAggregator) Stop() error {
	close(a.stopCh)
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
	// TODO: using workerGroup
	var wg sync.WaitGroup
	for symbol, ts := range symbolTrades {
		wg.Add(1)
		go func(sym string, ts []*Trade) {
			defer wg.Done()
			if value, ok := a.realtimeSymbolBars.Load(sym); ok {
				symbolBars := value.(*RealtimeSymbolBars)
				symbolBars.UpdateByTrades(ctx, ts)
			}
		}(symbol, ts)
	}
	wg.Wait()
}

// ==================== Interval Timer Management ====================

func (a *OHLCVAggregator) manageIntervalTimers(ctx context.Context) {
	for interval, config := range SupportedIntervals {
		go a.startIntervalTimer(ctx, interval, config)
	}
}

// startIntervalTimer process interval (1h, 1d, 1w, ...), if reached close bar time, do closeIntervalBars()
func (a *OHLCVAggregator) startIntervalTimer(ctx context.Context, interval OHLCV_INTERVAL, config IntervalConfig) {
	// Calculate next interval boundary
	now := time.Now()
	nextBoundary := getNextBucketTime(now, config.Duration)

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

func (a *OHLCVAggregator) closeIntervalBars(ctx context.Context, interval OHLCV_INTERVAL, openTime int64) {
	a.realtimeSymbolBars.Range(func(key, value interface{}) bool {
		rsBars := value.(*RealtimeSymbolBars)
		closedBars, err := rsBars.CloseBars(interval, openTime)
		if err = a.repo.SaveOHLCVBars(ctx, closedBars, interval); err != nil {
			log.Errorf("[OHLCVAggregator] Failed to save OHLCVBars: %v", err)
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
