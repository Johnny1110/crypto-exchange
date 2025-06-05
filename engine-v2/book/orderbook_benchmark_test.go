package book

import (
	"fmt"
	"github.com/johnny1110/crypto-exchange/engine-v2/market"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"testing"
)

// Benchmark for placing maker (limit) orders
func BenchmarkMakeLimitOrder(b *testing.B) {
	ob := NewOrderBook(mockMarket())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		order := model.NewOrder(
			fmt.Sprintf("M%08d", i),
			"bench-user",
			model.BID,
			float64(i%1000)+1000,
			1.0, 0,
			model.MAKER, 0,
		)
		if _, err := ob.PlaceOrder(model.LIMIT, order); err != nil {
			b.Fatalf("MakeLimitOrder failed: %v", err)
		}
	}
}

// Benchmark for taking limit orders that fully match at top of book
func BenchmarkTakeLimitOrder_FullMatch(b *testing.B) {
	// prepare book with deep book depth
	ob := NewOrderBook(mockMarket())
	depth := 1000
	for i := 0; i < depth; i++ {
		order := model.NewOrder(
			fmt.Sprintf("A%08d", i),
			"bench-user",
			model.ASK,
			1000+float64(i),
			1.0, 0,
			model.MAKER, 0,
		)
		ob.PlaceOrder(model.LIMIT, order)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		taker := model.NewOrder(
			fmt.Sprintf("T%08d", i),
			"bench-user",
			model.BID,
			1000+float64(i%depth),
			1.0, 0,
			model.TAKER, 0,
		)
		_, err := ob.PlaceOrder(model.LIMIT, taker)
		if err != nil {
			b.Fatalf("TakeLimitOrder full match failed: %v", err)
		}
	}
}

// Benchmark for taking market orders
func BenchmarkTakeMarketOrder(b *testing.B) {
	ob := NewOrderBook(mockMarket())
	depth := 5000
	volume := 0.0
	for i := 0; i < depth; i++ {
		size := 1.0
		volume += size
		order := model.NewOrder(
			fmt.Sprintf("A%08d", i),
			"bench-user",
			model.ASK,
			1000+float64(i),
			size, 0,
			model.MAKER, 0,
		)
		ob.PlaceOrder(model.LIMIT, order)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		taker := model.NewOrder(
			fmt.Sprintf("T%08d", i),
			"bench-user",
			model.BID,
			0,
			0.001, 0,
			model.TAKER, 0,
		)
		_, err := ob.PlaceOrder(model.MARKET, taker)
		if err != nil {
			b.Fatalf("TakeMarketOrder failed: %v", err)
		}
	}
}

// Benchmark for canceling orders
func BenchmarkCancelOrder(b *testing.B) {
	ob := NewOrderBook(mockMarket())
	// pre-insert orders
	orders := make([]*model.Order, b.N)
	for i := 0; i < b.N; i++ {
		orders[i] = model.NewOrder(
			fmt.Sprintf("C%08d", i),
			"bench-user",
			model.BID,
			1000,
			1.0, 0,
			model.MAKER, 0,
		)
		ob.PlaceOrder(model.LIMIT, orders[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ob.CancelOrder(orders[i].ID); err != nil {
			b.Fatalf("CancelOrder failed: %v", err)
		}
	}
}

func mockMarket() *market.MarketInfo {
	return market.NewMarketInfo("DOT/USDT", "DOT", "USDT")
}
