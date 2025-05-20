package book

import (
	"fmt"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"testing"
)

func mockOrderBook(t *testing.T) *OrderBook {
	ob := NewOrderBook("ETH/USDT")

	marketMakerName := "supermaker"
	// make some bid order (total 5 qty)
	bidOrder_1 := model.NewOrder("B01", marketMakerName, model.BID, 2100, 5)
	bidOrder_2 := model.NewOrder("B02", marketMakerName, model.BID, 2150, 5)
	bidOrder_3 := model.NewOrder("B03", marketMakerName, model.BID, 2200, 5)
	bidOrder_4 := model.NewOrder("B04", marketMakerName, model.BID, 2250, 5)
	bidOrder_5 := model.NewOrder("B05", marketMakerName, model.BID, 2300, 5)

	// make some ask order (total 5 qty)
	askOrder_1 := model.NewOrder("A01", marketMakerName, model.ASK, 2100, 4)
	askOrder_2 := model.NewOrder("A02", marketMakerName, model.ASK, 2150, 4)
	askOrder_3 := model.NewOrder("A03", marketMakerName, model.ASK, 2200, 4)
	askOrder_4 := model.NewOrder("A04", marketMakerName, model.ASK, 2250, 4)
	askOrder_5 := model.NewOrder("A05", marketMakerName, model.ASK, 2300, 4)

	ob.PlaceOrder(MAKE_LIMIT, bidOrder_1)
	ob.PlaceOrder(MAKE_LIMIT, bidOrder_2)
	ob.PlaceOrder(MAKE_LIMIT, bidOrder_3)
	ob.PlaceOrder(MAKE_LIMIT, bidOrder_4)
	ob.PlaceOrder(MAKE_LIMIT, bidOrder_5)

	ob.PlaceOrder(MAKE_LIMIT, askOrder_1)
	ob.PlaceOrder(MAKE_LIMIT, askOrder_2)
	ob.PlaceOrder(MAKE_LIMIT, askOrder_3)
	ob.PlaceOrder(MAKE_LIMIT, askOrder_4)
	ob.PlaceOrder(MAKE_LIMIT, askOrder_5)

	totalAsk := ob.TotalAskVolume()
	totalBid := ob.TotalBidVolume()
	assert(t, 25.0, totalBid)
	assert(t, 20.0, totalAsk)

	return ob
}

func TestOrderBook_AddSameOrderID(t *testing.T) {
	ob := mockOrderBook(t)
	bidOrder_1 := model.NewOrder("B01", "test01", model.BID, 2100, 5)
	var _, err_1 = ob.PlaceOrder(MAKE_LIMIT, bidOrder_1)
	fmt.Println(err_1)
	assert(t, true, err_1 != nil)

	var _, err_2 = ob.PlaceOrder(TAKE_LIMIT, bidOrder_1)
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
	bidOrder_qty1 := model.NewOrder("test_bid_01", "test01", model.BID, 2100, 1)
	trades, _ := ob.PlaceOrder(TAKE_LIMIT, bidOrder_qty1)

	fmt.Println(trades)
	assert(t, 1, len(trades))
	assert(t, 1.0, trades[0].Qty)
	assert(t, 2100.0, trades[0].Price)
	assert(t, "test_bid_01", trades[0].BidOrderID)
	assert(t, "A01", trades[0].AskOrderID)

	assert(t, 19.0, ob.TotalAskVolume())
	assert(t, 25.0, ob.TotalBidVolume())

	// buy 2150 can fill ask 2100 * 3 and 2150 * 4 & bid left 3 qty
	bidOrder_qty10 := model.NewOrder("test_bid_02", "test01", model.BID, 2150, 10)
	trades_2, _ := ob.PlaceOrder(TAKE_LIMIT, bidOrder_qty10)
	fmt.Println(trades_2)
	assert(t, 2, len(trades_2))
	assert(t, 25.0+3.0, ob.TotalBidVolume())

	// try add a same orderId
	bidOrder_qty10_same_id := model.NewOrder("test_bid_02", "test01", model.BID, 2150, 10)
	trades_3, err := ob.PlaceOrder(TAKE_LIMIT, bidOrder_qty10_same_id)
	assert(t, true, err != nil)
	fmt.Println(trades_3)
	fmt.Println(err)
}

func TestOrderBook_TakeMarketOrder(t *testing.T) {
	ob := mockOrderBook(t)
	fmt.Println(ob.TotalAskVolume())
	fmt.Println(ob.TotalBidVolume())

	askOrder_qty100 := model.NewOrder("test_ask_01", "test01", model.ASK, 0, 100)
	_, err := ob.PlaceOrder(MARKET, askOrder_qty100)
	assert(t, true, err != nil)
	fmt.Println(err)

	askOrder_qty10 := model.NewOrder("test_ask_01", "test01", model.ASK, 0, 11)
	trades, _ := ob.PlaceOrder(MARKET, askOrder_qty10)
	fmt.Println(trades)
	assert(t, 3, len(trades))
	assert(t, 5.0, trades[0].Qty)
	assert(t, 5.0, trades[1].Qty)
	assert(t, 1.0, trades[2].Qty)

	assert(t, 14.0, ob.TotalBidVolume())
}
