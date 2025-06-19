package ohlcv

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/labstack/gommon/log"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"math/rand"
	"testing"
	"time"
)

// ========================================== mock zone ==========================================
type mockRepo struct {
}

func (m mockRepo) SaveOHLCVBar(ctx context.Context, bar *OHLCVBar, interval OHLCV_INTERVAL) error {
	return nil
}

func (m mockRepo) GetOHLCVData(ctx context.Context, req *GetOhlcvDataReq) (*OHLCV, error) {
	return nil, nil
}

func (m mockRepo) UpdateRealtimeOHLCV(ctx context.Context, bar OHLCVBar, interval OHLCV_INTERVAL) error {
	log.Infof("* [Update] RealtimeOHLCV interval: %v, bar: %v", interval, bar)
	return nil
}

func (m mockRepo) GetRealtimeOHLCV(ctx context.Context, symbol, interval OHLCV_INTERVAL, openTime int64) (*OHLCVBar, error) {
	return nil, nil
}

func (m mockRepo) UpdateStatistics(ctx context.Context, symbol, interval OHLCV_INTERVAL, date time.Time, stats *ohlcvStatistics) error {
	return nil
}

func (m mockRepo) UpsertOHLCVBars(ctx context.Context, ohlcvBars []OHLCVBar, interval OHLCV_INTERVAL) error {
	log.Infof("* [Save] OHLCVBars: interval: %v >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>", interval)
	for _, bar := range ohlcvBars {
		log.Infof("OHLCVBars: %v", bar)
	}
	log.Infof("* [Save] OHLCVBars: interval: %v <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<", interval)
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

func mockAgg_with_SQLITE() (*OHLCVAggregator, error) {
	db, err := sql.Open("sqlite3", "../app/exg.db")
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := NewSQLiteOHLCVRepository(db)
	stream := createMockStream()
	return NewOHLCVAggregator(repo, stream, &AggregatorConfig{
		BatchSize:      10,
		FlushInterval:  3 * time.Second,
		ChannelSize:    10,
		MaxConcurrency: 2,
		EnableMetrics:  false,
	})
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
	agg, err := mockAgg_with_SQLITE()
	if err != nil {
		log.Errorf(err.Error())
		panic(err)
	}
	ctx := context.Background()
	err = agg.Start(ctx, []string{"ETH-USDT"})
	assert(t, err == nil, true)
	err = agg.AddSymbol("ETH-USDT", 2450, SupportedIntervals)
	assert(t, err == nil, true)

	testing_start_time := time.Now()

	for i := 0; i < 1000; i++ {
		time.Sleep(5 * time.Minute)

		ohlcv, err := agg.GetRealtimeOHLCV(ctx, "ETH-USDT", MIN_15)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("*** refresh realtime bar (15min): ", ohlcv)

		bars, err := agg.GetOHLCVData(ctx, &GetOhlcvDataReq{
			Symbol:    "ETH-USDT",
			Interval:  MIN_15,
			StartTime: testing_start_time,
			EndTime:   time.Now(),
		})

		if err != nil {
			t.Error(err)
		}
		fmt.Println("### refresh closed bar (15min)", bars)

		// ------------------------------------------------------------------------------------------------

		ohlcv, err = agg.GetRealtimeOHLCV(ctx, "ETH-USDT", H_1)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("*** refresh realtime bar (1h): ", ohlcv)

		bars, err = agg.GetOHLCVData(ctx, &GetOhlcvDataReq{
			Symbol:    "ETH-USDT",
			Interval:  H_1,
			StartTime: testing_start_time,
			EndTime:   time.Now(),
		})

		if err != nil {
			t.Error(err)
		}
		fmt.Println("### refresh closed bar (1h)", bars)

		// ------------------------------------------------------------------------------------------------

		ohlcv, err = agg.GetRealtimeOHLCV(ctx, "ETH-USDT", D_1)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("*** refresh realtime bar(1d): ", ohlcv)

		bars, err = agg.GetOHLCVData(ctx, &GetOhlcvDataReq{
			Symbol:    "ETH-USDT",
			Interval:  D_1,
			StartTime: testing_start_time,
			EndTime:   time.Now(),
		})

		if err != nil {
			t.Error(err)
		}
		fmt.Println("### refresh closed bar (1d)", bars)

		// ------------------------------------------------------------------------------------------------

		ohlcv, err = agg.GetRealtimeOHLCV(ctx, "ETH-USDT", W_1)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("*** refresh realtime bar(1w): ", ohlcv)

		bars, err = agg.GetOHLCVData(ctx, &GetOhlcvDataReq{
			Symbol:    "ETH-USDT",
			Interval:  W_1,
			StartTime: testing_start_time,
			EndTime:   time.Now(),
		})

		if err != nil {
			t.Error(err)
		}
		fmt.Println("### refresh closed bar (1w)", bars)

		fmt.Println("👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾👾")
	}
}

// ========================================== testing zone ==========================================
