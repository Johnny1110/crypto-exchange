package book

import (
	"fmt"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"sync"
	"testing"
	"time"
)

func mockOrderBook(t *testing.T) *OrderBook {
	ob := NewOrderBook(mockMarket())

	marketMakerName := "supermaker"
	// make some bid order (total 5 qty)
	bidOrder_1 := model.NewOrder("B01", marketMakerName, model.BID, 2100, 5, 0, model.MAKER)
	bidOrder_2 := model.NewOrder("B02", marketMakerName, model.BID, 2150, 5, 0, model.MAKER)
	bidOrder_3 := model.NewOrder("B03", marketMakerName, model.BID, 2200, 5, 0, model.MAKER)
	bidOrder_4 := model.NewOrder("B04", marketMakerName, model.BID, 2250, 5, 0, model.MAKER)
	bidOrder_5 := model.NewOrder("B05", marketMakerName, model.BID, 2300, 5, 0, model.MAKER)

	// make some ask order (total 5 qty)
	askOrder_1 := model.NewOrder("A01", marketMakerName, model.ASK, 2100, 4, 0, model.MAKER)
	askOrder_2 := model.NewOrder("A02", marketMakerName, model.ASK, 2150, 4, 0, model.MAKER)
	askOrder_3 := model.NewOrder("A03", marketMakerName, model.ASK, 2200, 4, 0, model.MAKER)
	askOrder_4 := model.NewOrder("A04", marketMakerName, model.ASK, 2250, 4, 0, model.MAKER)
	askOrder_5 := model.NewOrder("A05", marketMakerName, model.ASK, 2300, 4, 0, model.MAKER)

	ob.PlaceOrder(LIMIT, bidOrder_1)
	ob.PlaceOrder(LIMIT, bidOrder_2)
	ob.PlaceOrder(LIMIT, bidOrder_3)
	ob.PlaceOrder(LIMIT, bidOrder_4)
	ob.PlaceOrder(LIMIT, bidOrder_5)

	ob.PlaceOrder(LIMIT, askOrder_1)
	ob.PlaceOrder(LIMIT, askOrder_2)
	ob.PlaceOrder(LIMIT, askOrder_3)
	ob.PlaceOrder(LIMIT, askOrder_4)
	ob.PlaceOrder(LIMIT, askOrder_5)

	totalAsk := ob.TotalAskVolume()
	totalBid := ob.TotalBidVolume()
	assert(t, 25.0, totalBid)
	assert(t, 20.0, totalAsk)

	assert(t, ob.TotalBidQuoteAmount(), 55000.0)
	assert(t, ob.TotalAskQuoteAmount(), 44000.0)

	return ob
}

func TestOrderBook_AddSameOrderID(t *testing.T) {
	ob := mockOrderBook(t)
	bidOrder_1 := model.NewOrder("B01", "test01", model.BID, 2100, 5, 0, model.MAKER)
	var _, err_1 = ob.PlaceOrder(LIMIT, bidOrder_1)
	fmt.Println(err_1)
	assert(t, true, err_1 != nil)

	var _, err_2 = ob.PlaceOrder(LIMIT, bidOrder_1)
	fmt.Println(err_2)
	assert(t, true, err_2 != nil)

	var _, err_3 = ob.PlaceOrder(MARKET, bidOrder_1)
	fmt.Println(err_3)
	assert(t, true, err_3 != nil)
}

func TestOrderBook_MakeLimitOrder(t *testing.T) {
	ob := mockOrderBook(t)
	fmt.Println(ob.TotalAskVolume()) // 20
	fmt.Println(ob.TotalBidVolume()) // 25
}

func TestOrderBook_TakeLimitOrder_BID(t *testing.T) {
	// all ask volume in askSide is 20
	// price is from 2100 ~ 2300
	ob := mockOrderBook(t)
	bidOrder_qty1 := model.NewOrder("test_bid_01", "test01", model.BID, 2100, 1, 0, model.TAKER)
	trades, _ := ob.PlaceOrder(LIMIT, bidOrder_qty1)

	fmt.Println(trades)
	assert(t, 1, len(trades))
	assert(t, 1.0, trades[0].Size)
	assert(t, 2100.0, trades[0].Price)
	assert(t, "test_bid_01", trades[0].BidOrderID)
	assert(t, "A01", trades[0].AskOrderID)

	assert(t, 19.0, ob.TotalAskVolume())
	assert(t, 25.0, ob.TotalBidVolume())
	assert(t, ob.TotalBidQuoteAmount(), 55000.0)
	assert(t, ob.TotalAskQuoteAmount(), 41900.0)

	// buy 2150 can fill ask 2100 * 3 and 2150 * 4 & bid left 3 qty
	bidOrder_qty10 := model.NewOrder("test_bid_02", "test01", model.BID, 2150, 10, 0, model.TAKER)
	trades_2, _ := ob.PlaceOrder(LIMIT, bidOrder_qty10)
	fmt.Println(trades_2)
	assert(t, 2, len(trades_2))
	assert(t, 25.0+3.0, ob.TotalBidVolume())

	assert(t, ob.TotalBidQuoteAmount(), 55000.0+(3*2150))
	assert(t, ob.TotalAskQuoteAmount(), 41900.0-(2100*3+2150*4))

	// try add a same orderId
	bidOrder_qty10_same_id := model.NewOrder("test_bid_02", "test01", model.BID, 2150, 10, 0, model.TAKER)
	trades_3, err := ob.PlaceOrder(LIMIT, bidOrder_qty10_same_id)
	assert(t, true, err != nil)
	fmt.Println(trades_3)
	fmt.Println(err)

	// cancel bidOrder_qty10
	order, err := ob.CancelOrder("test_bid_02")
	assert(t, nil, err)
	fmt.Println("order", order)

	assert(t, order.RemainingSize, 3.0)
	assert(t, ob.TotalBidQuoteAmount(), 55000.0)
	assert(t, ob.TotalBidVolume(), 25.0)

}

func TestOrderBook_TakeMarketOrder(t *testing.T) {
	ob := mockOrderBook(t)
	fmt.Println(ob.TotalAskVolume())
	fmt.Println(ob.TotalBidVolume())

	askOrder_qty100 := model.NewOrder("test_ask_01", "test01", model.ASK, 0, 100, 0, model.TAKER)
	_, err := ob.PlaceOrder(MARKET, askOrder_qty100)
	assert(t, true, err != nil)
	fmt.Println(err)

	askOrder_qty10 := model.NewOrder("test_ask_01", "test01", model.ASK, 0, 11, 0, model.TAKER)
	trades, _ := ob.PlaceOrder(MARKET, askOrder_qty10)
	fmt.Println(trades)
	assert(t, 3, len(trades))
	assert(t, 5.0, trades[0].Size)
	assert(t, 5.0, trades[1].Size)
	assert(t, 1.0, trades[2].Size)

	assert(t, 14.0, ob.TotalBidVolume())
	fmt.Printf("Latest Price %.2f \n", ob.LatestPrice())
	assert(t, 2200.0, ob.LatestPrice())
}

// Helper to assert error absence
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Concurrency safety: spawn multiple goroutines for placing and canceling orders
func TestOrderBook_ConcurrencySafety(t *testing.T) {
	ob := NewOrderBook(mockMarket())
	const n = 1000
	var wg sync.WaitGroup

	// Place and then cancel orders concurrently
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("O%04d", i)
			order := model.NewOrder(id, "user", model.BID, 100+float64(i%10), 1, 0, model.MAKER)
			// Place order
			_, err_1 := ob.PlaceOrder(LIMIT, order)
			assertNoError(t, err_1)
			// Optional small sleep to increase interleaving
			time.Sleep(time.Microsecond)
			// Cancel order
			_, err := ob.CancelOrder(id)
			// Cancellation should succeed
			if err != nil {
				t.Errorf("cancel failed for %s: %v", id, err)
			}
		}(i)
	}
	wg.Wait()

	// After all, book should be empty
	assert(t, 0.0, ob.TotalBidVolume())
	assert(t, 0.0, ob.TotalAskVolume())
}

// Boundary scenarios
func TestOrderBook_BoundaryScenarios(t *testing.T) {
	ob := NewOrderBook(mockMarket())

	// Empty book matching returns no trades and no panics
	trades, err := ob.PlaceOrder(LIMIT, model.NewOrder("T1", "u", model.BID, 100, 1, 0, model.TAKER))
	assertNoError(t, err)
	assert(t, 0, len(trades))

	// Price mismatch: bid price lower than best ask
	// Setup an ask at price 110
	_, err_1 := ob.PlaceOrder(LIMIT, model.NewOrder("A1", "u", model.ASK, 110, 5, 0, model.MAKER))
	assertNoError(t, err_1)
	// Place a taker bid at price 100
	trades, err_2 := ob.PlaceOrder(LIMIT, model.NewOrder("T2", "u", model.BID, 100, 1, 0, model.TAKER))
	assertNoError(t, err_2)
	assert(t, 0, len(trades))

	// Partial fill should re-enter remainder
	// Place taker bid at price 110 for quantity 3 (ask has 5)
	trades, err = ob.PlaceOrder(LIMIT, model.NewOrder("T3", "u", model.BID, 110, 3, 0, model.TAKER))
	assertNoError(t, err)
	assert(t, 1, len(trades))
	assert(t, 3.0, trades[0].Size)
	// Remaining ask volume should be 2
	assert(t, 2.0, ob.TotalAskVolume())
	// Bid side got no volume (taker fully consumed) here.
	assert(t, 2.0, ob.TotalBidVolume())

	// Clean up
	_, err = ob.CancelOrder("A1")
	assertNoError(t, err)
}

// Market order tests
func TestOrderBook_MarketOrder(t *testing.T) {
	ob := NewOrderBook(mockMarket())

	// Setup depth: two asks totaling 5
	_, err_1 := ob.PlaceOrder(LIMIT, model.NewOrder("A1", "u", model.ASK, 100, 2, 0, model.MAKER))
	assertNoError(t, err_1)
	_, err_2 := ob.PlaceOrder(LIMIT, model.NewOrder("A2", "u", model.ASK, 101, 3, 0, model.MAKER))
	assertNoError(t, err_2)

	// Insufficient market buy order
	_, err := ob.PlaceOrder(MARKET, model.NewOrder("T1", "u", model.BID, 0, 10, 0, model.TAKER))
	if err == nil {
		t.Fatalf("expected error for insufficient volume, got nil")
	}

	// Sufficient market buy order
	trades, err := ob.PlaceOrder(MARKET, model.NewOrder("T2", "u", model.BID, 0, 5, 0, model.TAKER))
	assertNoError(t, err)
	// Should generate exactly 2 trades
	assert(t, 2, len(trades))
	// All asks consumed
	assert(t, 0.0, ob.TotalAskVolume())
}
