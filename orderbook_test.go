package main

import (
	"fmt"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected %v, got %v", b, a)
	}
}

func TestLimit(t *testing.T) {
	limit := NewLimit(10_000)
	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)
	buyOrderD := NewOrder(true, 2)

	limit.AddOrder(buyOrderA)
	limit.AddOrder(buyOrderB)
	limit.AddOrder(buyOrderC)
	limit.AddOrder(buyOrderD)

	limit.DeleteOrder(buyOrderB)

	fmt.Println(limit)
	fmt.Println(limit.Orders)
}

func TestPlaceLimitOrder(t *testing.T) {
	orderBook := NewOrderBook()

	sellOrderA := NewOrder(false, 5)
	sellOrderB := NewOrder(false, 5)
	sellOrderC := NewOrder(false, 10)

	orderBook.PlaceLimitOrder(10_000, sellOrderA)
	orderBook.PlaceLimitOrder(10_000, sellOrderB)
	orderBook.PlaceLimitOrder(11_000, sellOrderC)

	assert(t, len(orderBook.asks), 2)
}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderBook()

	sellOrder := NewOrder(false, 20)
	ob.PlaceLimitOrder(10_000, sellOrder)

	buyOrder := NewOrder(true, 10)
	matches := ob.PlaceMarketOrder(buyOrder)

	assert(t, len(matches), 1)
	assert(t, len(ob.asks), 1)
	assert(t, ob.AskTotalVolume(), 10.0)
	assert(t, matches[0].Ask, sellOrder)
	assert(t, matches[0].Bid, buyOrder)
	assert(t, matches[0].Price, 10_000.0)
	assert(t, matches[0].SizeFilled, 10.0)
	assert(t, buyOrder.IsFilled(), true)

	fmt.Printf("%+v\n", matches)
}

func TestPlaceMarketOrderMultiFill(t *testing.T) {
	ob := NewOrderBook()

	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)
	buyOrderD := NewOrder(true, 1)

	ob.PlaceLimitOrder(10_000, buyOrderA)
	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(5_000, buyOrderC)
	ob.PlaceLimitOrder(5_000, buyOrderD)

	assert(t, ob.BidTotalVolume(), 24.0)

	sellOrder := NewOrder(false, 20)
	matches := ob.PlaceMarketOrder(sellOrder)
	fmt.Printf("%+v\n", matches)
	assert(t, ob.BidTotalVolume(), 4.0)
	assert(t, len(matches), 3)

	fmt.Printf("the bids left: %+v\n", ob.bids)
	fmt.Printf("the bidLimits left: %+v\n", ob.BidLimits)
	assert(t, len(ob.bids), 1)
}
