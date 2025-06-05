package dto

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"time"
)

type Order struct {
	ID            string            `json:"id"`
	UserID        string            `json:"-"`
	Market        string            `json:"market"`
	Side          model.Side        `json:"side"`
	Price         float64           `json:"-"`
	OriginalSize  float64           `json:"original_size"`
	RemainingSize float64           `json:"remaining_size"`
	QuoteAmount   float64           `json:"quote_amount"`
	AvgDealtPrice float64           `json:"avg_dealt_price"`
	Type          model.OrderType   `json:"type"`
	Mode          model.Mode        `json:"mode"`
	Status        model.OrderStatus `json:"status"`
	FeeRate       float64           `json:"-"`
	Fees          float64           `json:"fees"`
	FeeAsset      string            `json:"fee_asset"`
	CreatedAt     time.Time         `json:"-"`
	UpdatedAt     time.Time         `json:"-"`
}

func (o Order) MarshalJSON() ([]byte, error) {
	type Alias Order

	// Create the struct with conditional price field
	result := struct {
		*Alias
		Price     *float64 `json:"price,omitempty"`
		FeeRate   string   `json:"fee_rate"`
		CreatedAt int64    `json:"created_at"`
		UpdatedAt int64    `json:"updated_at"`
	}{
		Alias:     (*Alias)(&o),
		FeeRate:   fmt.Sprintf("%.4f%%", o.FeeRate*100), // properly format percentage
		CreatedAt: o.CreatedAt.UnixMilli(),
		UpdatedAt: o.UpdatedAt.UnixMilli(),
	}

	// Only include price if it's > 0
	if o.Price > 0 {
		result.Price = &o.Price
	}

	return json.Marshal(result)
}

func (o Order) ToEngineOrder() *model.Order {
	return &model.Order{
		ID:            o.ID,
		UserID:        o.UserID,
		Side:          o.Side,
		Price:         o.Price,
		OriginalSize:  o.OriginalSize,
		RemainingSize: o.RemainingSize,
		Mode:          o.Mode,
		FeeRate:       o.FeeRate,
		Timestamp:     o.CreatedAt,
	}
}

// OrderBuilder provides a fluent interface for building orders
type OrderBuilder struct {
	order *Order
}

// NewOrderBuilder creates a new order builder
func NewOrderBuilder() *OrderBuilder {
	return &OrderBuilder{
		order: &Order{
			ID:        uuid.NewString(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

func (b *OrderBuilder) WithMarket(market string) *OrderBuilder {
	b.order.Market = market
	return b
}

func (b *OrderBuilder) WithUser(userID string) *OrderBuilder {
	b.order.UserID = userID
	return b
}

func (b *OrderBuilder) WithSide(side model.Side) *OrderBuilder {
	b.order.Side = side
	return b
}

func (b *OrderBuilder) WithType(orderType model.OrderType) *OrderBuilder {
	b.order.Type = orderType
	return b
}

func (b *OrderBuilder) WithMode(mode model.Mode) *OrderBuilder {
	b.order.Mode = mode
	return b
}

func (b *OrderBuilder) WithPrice(price float64) *OrderBuilder {
	b.order.Price = price
	return b
}

func (b *OrderBuilder) WithSize(size float64) *OrderBuilder {
	b.order.OriginalSize = size
	b.order.RemainingSize = size
	return b
}

func (b *OrderBuilder) WithQuoteAmount(amount float64) *OrderBuilder {
	b.order.QuoteAmount = amount
	return b
}

func (b *OrderBuilder) WithFeeRate(feeRate float64, feeAsset string) *OrderBuilder {
	b.order.FeeRate = feeRate
	b.order.FeeAsset = feeAsset
	return b
}

func (b *OrderBuilder) Build() *Order {
	b.order.Status = model.ORDER_STATUS_NEW
	return b.order
}

// PlaceOrderContext encapsulates all order placement context
type PlaceOrderContext struct {
	Market   string
	UserID   string
	Request  *OrderReq
	FeeRate  float64
	FeeAsset string
	OrderDTO *Order
	Assets   *AssetDetails
	Trades   []book.Trade
}

func (c *PlaceOrderContext) SyncTradeResult(engineOrder *model.Order, trades []book.Trade) {
	c.OrderDTO.RemainingSize = engineOrder.RemainingSize
	c.OrderDTO.Status = engineOrder.GetStatus()
	c.Trades = trades
}

// AssetDetails holds asset-related information
type AssetDetails struct {
	BaseAsset   string
	QuoteAsset  string
	FreezeAsset string
	FreezeAmt   float64
}
