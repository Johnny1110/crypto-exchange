package ohlcv

import (
	"fmt"
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
