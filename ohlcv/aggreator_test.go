package ohlcv

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// ========================================== mock zone ==========================================
type mockRepo struct {
}

func (m mockRepo) SaveOHLCVBar(ctx context.Context, bar *OHLCVBar, interval string) error {
	return nil
}

func (m mockRepo) GetOHLCVData(ctx context.Context, req *GetOhlcvDataReq) (*OHLCV, error) {
	return nil, nil
}

func (m mockRepo) UpdateRealtimeOHLCV(ctx context.Context, bar OHLCVBar, interval OHLCV_INTERVAL) error {
	return nil
}

func (m mockRepo) GetRealtimeOHLCV(ctx context.Context, symbol, interval string, openTime int64) (*OHLCVBar, error) {
	return nil, nil
}

func (m mockRepo) UpdateStatistics(ctx context.Context, symbol, interval string, date time.Time, stats *ohlcvStatistics) error {
	return nil
}

func (m mockRepo) SaveOHLCVBars(ctx context.Context, ohlcvBars []OHLCVBar, interval OHLCV_INTERVAL) error {
	return nil
}

type mockStream struct {
}

func (m mockStream) Subscribe(ctx context.Context, symbols []string) (<-chan *Trade, error) {
	// Check if ETH-USDT is in the requested symbols
	hasETHUSDT := false
	for _, symbol := range symbols {
		if symbol == "ETH-USDT" {
			hasETHUSDT = true
			break
		}
	}

	if !hasETHUSDT {
		return nil, fmt.Errorf("unsupported symbols: only ETH-USDT is supported")
	}

	tradeChan := make(chan *Trade, 1)

	go func() {
		defer close(tradeChan)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		basePrice := 2500.0

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Generate random price within ±1% of base price
				priceVariation := (rand.Float64()*2 - 1) * 0.01 // -1% to +1%
				price := basePrice * (1 + priceVariation)

				// Generate random volume between 0.01 and 0.1
				volume := 0.01 + rand.Float64()*0.09

				trade := &Trade{
					Symbol:    "ETH-USDT",
					Price:     price,
					Volume:    volume,
					Timestamp: time.Now(),
				}

				//log.Infof("[Test] mockStream Subscribe, sending teade data...")

				select {
				case tradeChan <- trade:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return tradeChan, nil
}

func (m mockStream) Close() error {
	return nil
}

func createMockRepo() OHLCVRepository {
	return &mockRepo{}
}

func createMockStream() TradeStream {
	return &mockStream{}
}

func mockAgg() (*OHLCVAggregator, error) {
	repo := createMockRepo()
	stream := createMockStream()
	return NewOHLCVAggregator(repo, stream, &AggregatorConfig{
		BatchSize:      10,
		FlushInterval:  3 * time.Second,
		ChannelSize:    10,
		MaxConcurrency: 2,
		EnableMetrics:  false,
	})
}

func Test_NewOHLCVAggregator(t *testing.T) {
	_, err := mockAgg()
	assert(t, err == nil, true)
}

func Test_Startup(t *testing.T) {
	agg, _ := mockAgg()
	ctx := context.Background()
	err := agg.Start(ctx, []string{"ETH-USDT"})
	assert(t, err == nil, true)
	err = agg.AddSymbol("ETH-USDT", 2450, SupportedIntervals)
	assert(t, err == nil, true)

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)

		ohlcv, err := agg.GetRealtimeOHLCV(ctx, "ETH-USDT", H_1)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("refresh bar: ", ohlcv)
	}
}

// ========================================== testing zone ==========================================
