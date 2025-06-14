package ohlcv

import (
	"fmt"
	"time"
)

// ========================= Config =========================

type IntervalConfig struct {
	Duration time.Duration
	Table    string
}

var SupportedIntervals = map[OHLCV_INTERVAL]IntervalConfig{
	H_1: {Duration: time.Hour, Table: "ohlcv_1h"},
	D_1: {Duration: 24 * time.Hour, Table: "ohlcv_1d"},
	W_1: {Duration: 7 * 24 * time.Hour, Table: "ohlcv_1w"},
}

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

type AggregatorConfig struct {
	BatchSize      int
	FlushInterval  time.Duration
	ChannelSize    int
	MaxConcurrency int
	EnableMetrics  bool
}

func (c *AggregatorConfig) Validate() error {
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

func DefaultAggregatorConfig() *AggregatorConfig {
	return &AggregatorConfig{
		BatchSize:      DefaultBatchSize,
		FlushInterval:  DefaultFlushInterval,
		ChannelSize:    DefaultChannelSize,
		MaxConcurrency: 10,
		EnableMetrics:  true,
	}
}
