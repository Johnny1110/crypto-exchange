package exchange

import (
	"encoding/json"
	"fmt"
	"github.com/johnny1110/crypto-exchange/orderbook"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

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
	Username string    `json:"username"`
	Market   Market    `json:"market"`
	Type     OrderType `json:"type"`
	Bid      bool      `json:"bid"`
	Size     float64   `json:"size"`
	Price    float64   `json:"price"`
}

func (ex *Exchange) InitOrderbooks() {
	ex.orderbooks[MarketBTC] = orderbook.NewOrderBook()
	ex.orderbooks[MarketETH] = orderbook.NewOrderBook()
}
func (ex *Exchange) HandlePlaceOrder(c echo.Context) error {

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

	switch orderType {
	case LimitOrder:
		if placeOrderData.Price != 0 {
			ob.PlaceLimitOrder(placeOrderData.Price, order)
			return c.JSON(http.StatusOK, map[string]any{"msg": "limit order placed"})
		} else {
			return c.JSON(http.StatusBadRequest, map[string]any{"msg": "missing price"})
		}
	case MarketOrder:
		matches := ob.PlaceMarketOrder(order)
		return c.JSON(http.StatusOK, map[string]any{"matches": "market order matches " + strconv.Itoa(len(matches))})
	default:
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "missing order type"})
	}
}

type Order struct {
	ID        int64
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
	Bid       bool    `json:"bid"`
	Timestamp int64   `json:"timestamp"`
}

type OrderBookDisplay struct {
	Asks            []*Order
	Bids            []*Order
	TotalAsksVolume float64
	TotalBidsVolume float64
}

func (ex *Exchange) HandleGetOrderBook(c echo.Context) error {
	//parse market from URL
	market := c.Param("market")
	ob, ok := ex.orderbooks[Market(market)]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "invalid market"})
	}
	fmt.Println("orderbook:", ob)
	//convert orderbook to DisplayData

	orderBookDisplay := OrderBookDisplay{
		Asks:            []*Order{},
		Bids:            []*Order{},
		TotalAsksVolume: ob.AskTotalVolume(),
		TotalBidsVolume: ob.BidTotalVolume(),
	}

	for _, asks := range ob.Asks() {
		for _, order := range asks.Orders {
			orderBookDisplay.Asks = append(orderBookDisplay.Asks, &Order{
				ID:        order.ID,
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
				ID:        order.ID,
				Price:     bids.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			})
		}

	}

	return c.JSON(http.StatusOK, orderBookDisplay)
}

func (ex *Exchange) HandleDeleteOrder(c echo.Context) error {
	marketStr := c.Param("market")
	orderIdStr := c.Param("id")

	market := Market(marketStr)
	orderId, _ := strconv.Atoi(orderIdStr)
	ob := ex.orderbooks[market]
	order := ob.GetOrderById(int64(orderId))
	ob.CancelOrder(order)

	return c.JSON(http.StatusOK, map[string]any{"msg": "OK"})
}

func (ex *Exchange) HandleGetOrderIds(c echo.Context) error {
	marketStr := c.Param("market")
	market := Market(marketStr)
	ob := ex.orderbooks[market]

	return c.JSON(http.StatusOK, map[string]any{"msg": ob.GetLimitOrderIds()})

}
