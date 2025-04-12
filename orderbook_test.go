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
