package book

import (
	"fmt"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected %v, got %v", b, a)
	}
}

// MockBookSide return mock bid and ask side
func MockBookSide() (*BookSide, *BookSide) {
	bidSide := NewBookSide(true)
	askSide := NewBookSide(false)

	userId := "U_01"

	// create bid
	bidOrderNode_1 := model.NewOrderNode("1", userId, model.BID, 1000, 10, 0, model.MAKER, 0)
	bidOrderNode_2 := model.NewOrderNode("2", userId, model.BID, 1000, 15, 0, model.MAKER, 0)
	bidSide.AddOrderNode(1000, bidOrderNode_1)
	bidSide.AddOrderNode(1000, bidOrderNode_2)
	bidOrderNode_3 := model.NewOrderNode("3", userId, model.BID, 1200, 3, 0, model.MAKER, 0)
	bidSide.AddOrderNode(1200, bidOrderNode_3)
	bidOrderNode_4 := model.NewOrderNode("3", userId, model.BID, 1300, 2, 0, model.MAKER, 0)
	bidSide.AddOrderNode(1300, bidOrderNode_4)

	askOrderNode_1 := model.NewOrderNode("1", userId, model.BID, 1400, 10, 0, model.MAKER, 0)
	askOrderNode_2 := model.NewOrderNode("2", userId, model.BID, 1400, 15, 0, model.MAKER, 0)
	askSide.AddOrderNode(1400, askOrderNode_1)
	askSide.AddOrderNode(1400, askOrderNode_2)
	askOrderNode_3 := model.NewOrderNode("3", userId, model.BID, 1500, 3, 0, model.MAKER, 0)
	askSide.AddOrderNode(1500, askOrderNode_3)
	askOrderNode_4 := model.NewOrderNode("3", userId, model.BID, 1550, 2, 0, model.MAKER, 0)
	askSide.AddOrderNode(1550, askOrderNode_4)

	return bidSide, askSide
}

func TestAddOrderNode(t *testing.T) {
	bidSide, askSide := MockBookSide()
	assert(t, 3, bidSide.Len())
	assert(t, 3, askSide.Len())
}

func TestAddThenRemoveOrderNode(t *testing.T) {
	bidSide, _ := MockBookSide()
	askOrderNode_1 := model.NewOrderNode("777", "1", model.BID, 1400, 5, 0, model.MAKER, 0)
	bidSide.AddOrderNode(1400, askOrderNode_1)

	fmt.Println(bidSide.Len())
	assert(t, 4, bidSide.Len())
	assert(t, 35.0, bidSide.totalVolume)
	assert(t, 38200.0, bidSide.totalQuoteAmount)

	bidSide.RemoveOrderNode(1400, askOrderNode_1)
	fmt.Println(bidSide.Len())
	assert(t, 3, bidSide.Len())
	assert(t, 30.0, bidSide.totalVolume)
	assert(t, 31200.0, bidSide.totalQuoteAmount)

	fmt.Println(bidSide.TotalVolume())
	fmt.Println(bidSide.TotalQuoteAmount())
}

func TestBestPrice(t *testing.T) {
	bidSide, askSide := MockBookSide()

	bestBIdPrice, _ := bidSide.BestPrice()
	assert(t, 1300.0, bestBIdPrice)

	bestAskPrice, _ := askSide.BestPrice()
	assert(t, 1400.0, bestAskPrice)
}

func TestPopBest(t *testing.T) {
	bidSide, askSide := MockBookSide()

	bestBidOrder, _ := bidSide.PopBest()
	assert(t, 1300.0, bestBidOrder.Price())
	assert(t, 2.0, bestBidOrder.Size())

	bestAskOrder, _ := askSide.PopBest()
	assert(t, 1400.0, bestAskOrder.Price())
	assert(t, 10.0, bestAskOrder.Size())
}
