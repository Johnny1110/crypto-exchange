package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/johnny1110/crypto-exchange/orderbook"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	ex := NewExchange()
	ex.initOrderbooks()
	e.POST("/order", ex.handlePlaceOrder)
	e.GET("/orderbook/:market", ex.handleGetOrderBook)

	e.Start(":3000")
}

type Market string

const (
	MarketBTC Market = "BTC"
	MarketETH Market = "ETH"
)

type Exchange struct {
	orderbooks map[Market]*orderbook.OrderBook
}

func NewExchange() *Exchange {
	return &Exchange{
		orderbooks: make(map[Market]*orderbook.OrderBook),
	}
}

type OrderType string

const (
	LimitOrder  OrderType = "LIMIT"
	MarketOrder OrderType = "MARKET"
)

type PlaceOrderRequest struct {
	Market Market    `json:"market"`
	Type   OrderType `json:"type"`
	Bid    bool      `json:"bid"`
	Size   float64   `json:"size"`
	Price  float64   `json:"price"`
}

func (ex *Exchange) initOrderbooks() {
	ex.orderbooks[MarketBTC] = orderbook.NewOrderBook()
	ex.orderbooks[MarketETH] = orderbook.NewOrderBook()
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {

	var placeOrderData PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return c.JSON(400, "{'error': 'invalid request'}")
	}

	market := Market(placeOrderData.Market)
	orderType := OrderType(placeOrderData.Type)
	ob := ex.orderbooks[market]

	if ob == nil {
		return c.JSON(400, "{'error': 'invalid market'}")
	}

	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size)

	if (orderType == LimitOrder) && (placeOrderData.Price != 0) {
		// process limit order
		ob.PlaceLimitOrder(placeOrderData.Price, order)
	}
	if orderType == MarketOrder {
		if err := ob.PlaceMarketOrder(order); err != nil {
			return c.JSON(400, fmt.Sprintf("{'error': '%s'}", err))
		}

	}

	return c.JSON(200, "{'msg': 'order placed'}")
}

type Order struct {
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
	Bid       bool    `json:"bid"`
	Timestamp int64   `json:"timestamp"`
}

type OrderBookDisplay struct {
	Asks []*Order
	Bids []*Order
}

func (ex *Exchange) handleGetOrderBook(c echo.Context) error {
	//parse market from URL
	market := c.Param("market")
	ob, ok := ex.orderbooks[Market(market)]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "invalid market"})
	}
	fmt.Println("orderbook:", ob)
	//convert orderbook to DisplayData

	orderBookDisplay := OrderBookDisplay{
		Asks: []*Order{},
		Bids: []*Order{},
	}

	for _, asks := range ob.Asks() {
		for _, order := range asks.Orders {
			orderBookDisplay.Asks = append(orderBookDisplay.Asks, &Order{
				Price:     asks.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			})
		}
	}

	for _, bids := range ob.Bids() {
		for _, order := range bids.Orders {
			orderBookDisplay.Bids = append(orderBookDisplay.Bids, &Order{
				Price:     bids.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			})
		}

	}

	return c.JSON(http.StatusAccepted, orderBookDisplay)
}
