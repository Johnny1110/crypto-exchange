package ohlcv

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func assert(t *testing.T, actual, expected any) {
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func Test_NewRealtimeSymbolBars(t *testing.T) {
	rtSymbolBars := mockSymbolBars()
	assert(t, rtSymbolBars.symbol, "ETH-USDT")
	fmt.Println(rtSymbolBars.GetAllIntervals())
	assert(t, 3, len(rtSymbolBars.GetAllIntervals()))
}

func mockSymbolBars() *RealtimeSymbolBars {
	return NewRealtimeSymbolBars("ETH-USDT", 0.01, SupportedIntervals)
}

func mockTradesData_simple5(mockCurrentTime time.Time) []*Trade {
	trades := make([]*Trade, 0, 5)

	trades = append(trades, &Trade{
		Symbol:    "ETH-USDT",
		Price:     2501.0,
		Volume:    0.1,
		Timestamp: mockCurrentTime,
	})

	trades = append(trades, &Trade{
		Symbol:    "ETH-USDT",
		Price:     2502.0,
		Volume:    0.1,
		Timestamp: mockCurrentTime.Add(1 * time.Second),
	})

	trades = append(trades, &Trade{
		Symbol:    "ETH-USDT",
		Price:     2505.0,
		Volume:    0.1,
		Timestamp: mockCurrentTime.Add(2 * time.Second),
	})

	trades = append(trades, &Trade{
		Symbol:    "ETH-USDT",
		Price:     2499.0,
		Volume:    0.1,
		Timestamp: mockCurrentTime.Add(3 * time.Second),
	})

	trades = append(trades, &Trade{
		Symbol:    "ETH-USDT",
		Price:     2503.0,
		Volume:    0.1,
		Timestamp: mockCurrentTime.Add(4 * time.Second),
	})

	return trades
}

func mockTradesData_1h_AcrossHours(mockCurrentTime time.Time) []*Trade {
	trades := make([]*Trade, 0, 10)
	current_5 := mockTradesData_simple5(mockCurrentTime)
	trades = append(trades, current_5...)

	nextHour := mockCurrentTime.Add(1 * time.Hour)

	trades = append(trades, &Trade{
		Symbol:    "ETH-USDT",
		Price:     2601.0,
		Volume:    0.1,
		Timestamp: nextHour,
	})

	trades = append(trades, &Trade{
		Symbol:    "ETH-USDT",
		Price:     2602.0,
		Volume:    0.1,
		Timestamp: nextHour.Add(1 * time.Second),
	})

	trades = append(trades, &Trade{
		Symbol:    "ETH-USDT",
		Price:     2603.0,
		Volume:    0.1,
		Timestamp: nextHour.Add(2 * time.Second),
	})

	trades = append(trades, &Trade{
		Symbol:    "ETH-USDT",
		Price:     2590.0,
		Volume:    0.1,
		Timestamp: nextHour.Add(3 * time.Second),
	})

	trades = append(trades, &Trade{
		Symbol:    "ETH-USDT",
		Price:     2610.0,
		Volume:    0.1,
		Timestamp: nextHour.Add(4 * time.Second),
	})

	return trades
}

func Test_UpdateSymbolBars(t *testing.T) {
	rtSymbolBars := mockSymbolBars()
	ctx := context.Background()
	trades := mockTradesData_simple5(time.Now())
	rtSymbolBars.UpdateByTrades(ctx, trades)

	h1_bar, ok := rtSymbolBars.GetIntervalBar(H_1)
	assert(t, true, ok)
	fmt.Println("h1_bar", h1_bar)
	fmt.Println(rtSymbolBars.intervalBars[H_1])

	d1_bar, ok := rtSymbolBars.GetIntervalBar(D_1)
	assert(t, true, ok)
	fmt.Println("d1_bar", d1_bar)
	fmt.Println(rtSymbolBars.intervalBars[D_1])

	w1_bar, ok := rtSymbolBars.GetIntervalBar(W_1)
	assert(t, true, ok)
	fmt.Println("w1_bar", w1_bar)
	fmt.Println(rtSymbolBars.intervalBars[W_1])
}

func Test_UpdateSymbolBars_1h(t *testing.T) {
	rtSymbolBars := mockSymbolBars()
	ctx := context.Background()

	now := time.Now()

	currentH1Bucket := getBucketUnixTime(now, 1*time.Hour)
	nextH1Bucket := getNextBucketUnixTime(now, 1*time.Hour)

	trades := mockTradesData_1h_AcrossHours(now)
	rtSymbolBars.UpdateByTrades(ctx, trades)

	h1_bar, ok := rtSymbolBars.GetIntervalBar(H_1)
	assert(t, true, ok)
	fmt.Println("h1_bar", h1_bar)

	m, ok := rtSymbolBars.intervalBars[H_1]
	assert(t, ok, true)

	currentBar, ok := m[currentH1Bucket]
	assert(t, ok, true)
	nextBar, ok := m[nextH1Bucket]
	assert(t, ok, true)

	fmt.Println("currentBar", currentBar)
	assert(t, currentBar.OpenPrice, 0.01)
	assert(t, currentBar.HighPrice, 2505.0)
	assert(t, currentBar.LowPrice, 0.01)
	assert(t, currentBar.ClosePrice, 2503.0)
	assert(t, currentBar.Volume, 0.5)

	fmt.Println("nextBar", nextBar)
	assert(t, nextBar.OpenPrice, 2601.0)
	assert(t, nextBar.HighPrice, 2610.0)
	assert(t, nextBar.LowPrice, 2590.0)
	assert(t, nextBar.ClosePrice, 2610.0)
	assert(t, nextBar.Volume, 0.5)
}

func Test_CloseBars_AlreadyHaveNewLatest(t *testing.T) {
	rtSymbolBars := mockSymbolBars()
	ctx := context.Background()

	now := time.Now()

	currentH1Bucket := getBucketUnixTime(now, 1*time.Hour)

	trades := mockTradesData_1h_AcrossHours(now)
	rtSymbolBars.UpdateByTrades(ctx, trades)

	h1_bar, ok := rtSymbolBars.GetIntervalBar(H_1)
	assert(t, true, ok)
	fmt.Println("h1_bar", h1_bar)

	closedBars, err := rtSymbolBars.CloseBars(H_1, currentH1Bucket)
	assert(t, nil, err)
	fmt.Println("closedBars", closedBars)

	fmt.Println(rtSymbolBars.intervalBars[H_1])
	assert(t, 1, len(rtSymbolBars.intervalBars[H_1]))

	nextH1Bucket := getNextBucketUnixTime(now, 1*time.Hour)
	fmt.Println(rtSymbolBars.intervalBars[H_1][nextH1Bucket])

	assert(t, 1, len(rtSymbolBars.intervalBars[H_1]))
}

func Test_CloseBars_DontHaveNewLatest(t *testing.T) {
	rtSymbolBars := mockSymbolBars()
	ctx := context.Background()

	now := time.Now()

	currentH1Bucket := getBucketUnixTime(now, 1*time.Hour)

	trades := mockTradesData_simple5(now)
	rtSymbolBars.UpdateByTrades(ctx, trades)

	h1_bar, ok := rtSymbolBars.GetIntervalBar(H_1)
	assert(t, true, ok)
	fmt.Println("h1_bar", h1_bar)

	closedBars, err := rtSymbolBars.CloseBars(H_1, currentH1Bucket)
	assert(t, nil, err)
	fmt.Println("closedBars", closedBars)

	fmt.Println(rtSymbolBars.intervalBars[H_1])
	assert(t, 1, len(rtSymbolBars.intervalBars[H_1]))

	nextH1Bucket := getNextBucketUnixTime(now, 1*time.Hour)
	fmt.Println(rtSymbolBars.intervalBars[H_1][nextH1Bucket])
	newBar := rtSymbolBars.intervalBars[H_1][nextH1Bucket]

	assert(t, newBar.OpenPrice, closedBars[0].ClosePrice)
}
