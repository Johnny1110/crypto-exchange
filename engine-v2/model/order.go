package model

import (
	"github.com/johnny1110/crypto-exchange/utils"
	"time"
)

type Side int
type Mode int
type OrderType int
type OrderStatus string

const (
	LIMIT OrderType = iota
	MARKET
)

const (
	BID Side = iota
	ASK
)

const (
	MAKER Mode = iota
	TAKER
)

const (
	// ORDER_STATUS_NEW indicates an order that has just been created and not yet matched.
	ORDER_STATUS_NEW OrderStatus = "NEW"

	// ORDER_STATUS_PARTIAL indicates an order that has been partially filled.
	ORDER_STATUS_PARTIAL OrderStatus = "PARTIAL"

	// ORDER_STATUS_FILLED indicates an order that has been completely filled.
	ORDER_STATUS_FILLED OrderStatus = "FILLED"

	// ORDER_STATUS_CANCELED indicates an order that has been canceled.
	ORDER_STATUS_CANCELED OrderStatus = "CANCELED"
)

func (ot OrderType) String() string {
	switch ot {
	case LIMIT:
		return "LIMIT"
	case MARKET:
		return "MARKET"
	default:
		return "UNKNOWN"
	}
}

func (si Side) String() string {
	switch si {
	case BID:
		return "BID"
	case ASK:
		return "ASK"
	default:
		return "UNKNOWN"
	}
}

func (mo Mode) String() string {
	switch mo {
	case MAKER:
		return "MAKER"
	case TAKER:
		return "TAKER"
	default:
		return "UNKNOWN"
	}
}

type Order struct {
	ID            string
	UserID        string
	Side          Side
	Price         float64
	OriginalSize  float64
	RemainingSize float64
	QuoteAmount   float64 // only market bid order
	Mode          Mode
	FeeRate       float64
	Timestamp     time.Time
}

func (o *Order) GetStatus() OrderStatus {
	if o.OriginalSize == o.RemainingSize {
		return ORDER_STATUS_NEW
	}
	if o.RemainingSize <= utils.Scale {
		return ORDER_STATUS_FILLED
	}
	if o.RemainingSize > 0 && o.RemainingSize < o.OriginalSize {
		return ORDER_STATUS_PARTIAL
	}
	return ORDER_STATUS_CANCELED
}

func (o *Order) CounterSide() Side {
	return o.Side ^ 1
}

// NewOrder
// side: BID ASK
// mode: MAKER TAKER
func NewOrder(orderId, userId string, side Side, price float64, size float64, quoteAmt float64, mode Mode, feeRate float64) *Order {
	return &Order{
		ID:            orderId,
		UserID:        userId,
		Side:          side,
		Price:         price,
		OriginalSize:  size,
		RemainingSize: size,
		QuoteAmount:   quoteAmt,
		FeeRate:       feeRate,
		Mode:          mode,
		Timestamp:     time.Now(),
	}
}

type OrderNode struct {
	Order      *Order
	Prev, Next *OrderNode
}

func NewOrderNode(orderId, userId string, side Side, price float64, size float64, quoteAmt float64, orderType Mode, feeRate float64) *OrderNode {
	order := NewOrder(orderId, userId, side, price, size, quoteAmt, orderType, feeRate)
	return &OrderNode{
		Order: order,
	}
}

func (node *OrderNode) Size() float64 {
	return node.Order.RemainingSize
}

func (node *OrderNode) Price() float64 {
	return node.Order.Price
}
