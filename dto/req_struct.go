package dto

import (
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
)

type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SettlementReq struct {
	Username string  `json:"username" binding:"required"`
	Asset    string  `json:"asset" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
}

type OrderReq struct {
	Side        model.Side      `json:"side" binding:"oneof=0 1"`                          // 0=Bid,1=Ask
	OrderType   model.OrderType `json:"order_type" binding:"oneof=0 1"`                    // 0=LIMIT,1=MARKET
	Mode        model.Mode      `json:"mode" binding:"required_if=order_type 0,oneof=0 1"` // 0=MAKER,1=TAKER
	Price       float64         `json:"price"`                                             // only LIMIT order, and > 0
	Size        float64         `json:"size"`                                              // only market bid no need
	QuoteAmount float64         `json:"quote_amount"`                                      // only for taker bid order
}

type OrdersQueryType = string

const (
	OPENING_ORDER = OrdersQueryType("OPENING")
	CLOSED_ORDER  = OrdersQueryType("CLOSED")
)

type GetOrdersQueryReq struct {
	UserID      string
	Market      string          `form:"market"`
	Side        model.Side      `form:"side"`
	Type        OrdersQueryType `form:"type" binding:"required"`
	PageSize    int64           `form:"page_size,default=10"`
	CurrentPage int64           `form:"current_page,default=1"`
}
