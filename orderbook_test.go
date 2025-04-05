package main

import (
	"fmt"
	"testing"
)

func TestLimit(t *testing.T) {
	limit := NewLimit(10_000)
	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)

	limit.AddOrder(buyOrderA)
	limit.AddOrder(buyOrderB)
	limit.AddOrder(buyOrderC)

	limit.DeleteOrder(buyOrderB)

	fmt.Println(limit)
}

func TestOrderBook(t *testing.T) {

}
