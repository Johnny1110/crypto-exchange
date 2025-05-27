package book

import (
	"errors"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
)

// indexEntry stores the side, price level, and pointer to the order node for quick lookup.
type indexEntry struct {
	Side  model.Side
	Price float64
	Node  *model.OrderNode
}

// OrderIndex maintains a mapping from order ID to its indexEntry.
type OrderIndex struct {
	index map[string]indexEntry
}

func NewOrderIndex() *OrderIndex {
	return &OrderIndex{index: make(map[string]indexEntry)}
}

// Add inserts a new mapping for the given order ID.
func (oi *OrderIndex) Add(node *model.OrderNode) {
	order := node.Order
	oi.index[order.ID] = indexEntry{Side: order.Side, Price: order.Price, Node: node}
}

// Remove deletes the mapping for the given order ID.
// Returns an order, error if the order ID is not found.
func (oi *OrderIndex) Remove(orderID string) (*model.Order, error) {
	entry, found := oi.index[orderID]
	if !found {
		return nil, errors.New("order ID not found in index")
	}
	order := entry.Node.Order
	delete(oi.index, orderID)
	return order, nil
}

// Get retrieves the indexEntry for the given order ID.
// Returns the entry and true if found, or an empty entry and false otherwise.
func (oi *OrderIndex) Get(orderID string) (model.Side, float64, *model.OrderNode, bool) {
	entry, found := oi.index[orderID]
	if !found {
		return 0, 0, nil, false
	}
	return entry.Side, entry.Price, entry.Node, true
}

func (oi *OrderIndex) OrderIdExist(id string) bool {
	_, found := oi.index[id]
	return found
}
