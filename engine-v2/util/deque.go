package util

import (
	"errors"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
)

// OrderNodeDeque is a double end queue implement with double linked list.
// It's holds pointers of OrderNode defined in entity pkg.
// User must create OrderNode via model.OrderNode and push them into OrderNodeDeque.
type OrderNodeDeque struct {
	head, tail *model.OrderNode
	size       int
	volume     float64
}

func NewOrderNodeDeque() *OrderNodeDeque {
	return &OrderNodeDeque{
		size:   0,
		volume: 0,
	}
}

// PushBack adds a node to the end of the deque in O(1) time.
func (dq *OrderNodeDeque) PushBack(node *model.OrderNode) {
	if dq.tail == nil {
		dq.head = node
		dq.tail = node
	} else {
		node.Next = nil
		node.Prev = dq.tail
		dq.tail.Next = node
		dq.tail = node
	}
	dq.size++
	dq.volume += node.Size()
}

// PushHead adds a node to the head of the deque in O(1) time.
func (dq *OrderNodeDeque) PushHead(node *model.OrderNode) {
	if dq.head == nil {
		dq.head = node
		dq.tail = node
	} else {
		node.Prev = nil
		node.Next = dq.head
		dq.head.Prev = node
		dq.head = node
	}
	dq.size++
	dq.volume += node.Size()
}

// PopFront removes and returns the node at the front of the deque in O(1) time.
// Returns nil and an error if the deque is empty.
func (dq *OrderNodeDeque) PopFront() (*model.OrderNode, error) {
	if dq.head == nil {
		return nil, errors.New("OrderNodeDeque is empty")
	}
	node := dq.head

	if dq.size == 1 {
		dq.head = nil
		dq.tail = nil
	} else {
		dq.head = dq.head.Next
		dq.head.Prev = nil
	}
	dq.size--
	dq.volume -= node.Size()

	// clean up the popped node.
	node.Next = nil
	node.Prev = nil // should be nil already, but do it again for safety.
	return node, nil
}

// PeekFront returns the node at the front without removing it.
// Returns nil if the deque is empty.
func (dq *OrderNodeDeque) PeekFront() *model.OrderNode {
	return dq.head
}

// Remove removes a specific node from the deque in O(1) time.
// Returns an error if the node is nil or deque is empty.
func (dq *OrderNodeDeque) Remove(node *model.OrderNode) error {
	if node == nil {
		return errors.New("input node is nil")
	}
	if dq.size == 0 {
		return errors.New("OrderNodeDeque is empty")
	}

	// If node is head
	if dq.head == node {
		_, err := dq.PopFront()
		return err
	}

	// if node is tail
	if dq.tail == node {
		dq.tail = node.Prev
		dq.tail.Next = nil
		node.Prev = nil
		dq.size--
		dq.volume -= node.Size()
		return nil
	}

	// Node is in middle
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
	node.Next = nil
	node.Prev = nil
	dq.size--
	dq.volume -= node.Size()
	return nil
}

// IsEmpty returns true if the deque has no elements.
func (d *OrderNodeDeque) IsEmpty() bool {
	return d.size == 0
}

// Size returns the number of elements in the deque.
func (d *OrderNodeDeque) Size() int {
	return d.size
}

// Volume returns the number of elements in the deque.
func (d *OrderNodeDeque) Volume() float64 {
	return d.volume
}
