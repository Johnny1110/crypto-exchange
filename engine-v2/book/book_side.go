package book

import (
	"errors"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/engine-v2/util"
)

// BookSide represents one side (bid or ask) of the order book
// internally maintained as an ordered map from price -> deque of orders.
// For ask, highest price has priority; for bid, lowest price.
// The comparison function depends on the side.
type BookSide struct {
	priceLevels *treemap.Map // key: float64 price, value: *util.Deque (price ordered map)
	isBid       bool         // true = bid side (min-first), false = ask side (max-first)
	totalVolume float64      // all volume sit in bookSide
}

func NewBookSide(isBid bool) *BookSide {
	// choose comparator: reverse for buys
	var cmp utils.Comparator
	if isBid {
		cmp = utils.Float64Comparator // will treat smaller < larger, but we'll always call Rightmost for buys
	} else {
		cmp = utils.Float64Comparator // same comparator
	}
	return &BookSide{
		priceLevels: treemap.NewWith(cmp),
		isBid:       isBid,
		totalVolume: 0,
	}
}

// AddOrderNode inserts a node at a given price level, creating the level if needed.
func (bs *BookSide) AddOrderNode(price float64, node *model.OrderNode) {
	v, ok := bs.priceLevels.Get(price)
	if !ok {
		deque := util.NewOrderNodeDeque()
		bs.priceLevels.Put(price, deque)
		v = deque
	}
	// force convert to type OrderNodeDeque
	deque := v.(*util.OrderNodeDeque)
	deque.PushBack(node)
	bs.totalVolume += node.Qty()
}

// RemoveOrderNode removes a specific node from the deque at price.
// If the deque becomes empty, removes the price level.
func (bs *BookSide) RemoveOrderNode(price float64, node *model.OrderNode) error {
	v, ok := bs.priceLevels.Get(price)
	if !ok {
		errors.New("price level not found")
	}

	deque := v.(*util.OrderNodeDeque)
	err := deque.Remove(node)
	if err != nil {
		return err
	}

	// remove price level if current price level is empty
	if deque.IsEmpty() {
		bs.priceLevels.Remove(price)
	}

	bs.totalVolume -= node.Qty()
	return nil
}

// BestPrice returns the best price on this side (max for buys, min for sells).
func (bs *BookSide) BestPrice() (float64, error) {
	if bs.priceLevels.Empty() {
		return 0, errors.New("no price levels sit in book side")
	}

	// Bid is buy side
	if bs.isBid {
		// best bid(buy) price is highest price
		k, _ := bs.priceLevels.Max()
		return k.(float64), nil
	} else {
		// Ask is sell side, best ask(sell) price is lowest price
		k, _ := bs.priceLevels.Min()
		return k.(float64), nil
	}
}

// PeekBest returns the earliest OrderNode at the best price without removing it.
func (bs *BookSide) PeekBest() (*model.OrderNode, error) {
	bestPrice, err := bs.BestPrice()
	if err != nil {
		return nil, err
	}

	dq, _ := bs.priceLevels.Get(bestPrice)
	return dq.(*util.OrderNodeDeque).PeekFront(), nil
}

// PopBest removes and returns the earliest OrderNode at the best price.
// Also cleans up the price level if empty.
func (bs *BookSide) PopBest() (*model.OrderNode, error) {
	bestPrice, err := bs.BestPrice()
	if err != nil {
		return nil, err
	}
	v, _ := bs.priceLevels.Get(bestPrice)
	deque := v.(*util.OrderNodeDeque)
	node, err := deque.PopFront()
	if err != nil {
		return nil, err
	}
	if deque.IsEmpty() {
		bs.priceLevels.Remove(bestPrice)
	}

	bs.totalVolume -= node.Qty()
	return node, nil
}

// HasPriceLevel checks if a given price level exists.
func (bs *BookSide) HasPriceLevel(price float64) bool {
	_, found := bs.priceLevels.Get(price)
	return found
}

// Len returns number of price levels.
func (bs *BookSide) Len() int {
	return bs.priceLevels.Size()
}

func (bs *BookSide) TotalVolume() float64 {
	return bs.totalVolume
}

func (bs *BookSide) PutToHead(price float64, node *model.OrderNode) {
	v, ok := bs.priceLevels.Get(price)
	if !ok {
		deque := util.NewOrderNodeDeque()
		bs.priceLevels.Put(price, deque)
		v = deque
	}
	// force convert to type OrderNodeDeque
	deque := v.(*util.OrderNodeDeque)
	deque.PushHead(node)
	bs.totalVolume += node.Qty()
}
