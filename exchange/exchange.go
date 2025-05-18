package exchange

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/johnny1110/crypto-exchange/user"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/johnny1110/crypto-exchange/chainUtil"
	"github.com/johnny1110/crypto-exchange/orderbook"
	"github.com/labstack/echo/v4"
)

type (
	Market    string
	OrderType string

	Exchange struct {
		// users - key: orderId
		userIdMapUser    map[int64]*user.User
		orderIdMapUserId map[int64]int64
		orderbooks       map[Market]*orderbook.OrderBook
		address          string
		privateKey       *ecdsa.PrivateKey

		ethClient *ethclient.Client
	}

	PlaceOrderRequest struct {
		UserID   int64     `json:"userId"`
		Username string    `json:"username"`
		Market   Market    `json:"market"`
		Type     OrderType `json:"type"`
		Bid      bool      `json:"bid"`
		Size     float64   `json:"size"`
		Price    float64   `json:"price"`
	}

	Order struct {
		ID        int64
		Price     float64 `json:"price"`
		Size      float64 `json:"size"`
		Bid       bool    `json:"bid"`
		Timestamp int64   `json:"timestamp"`
	}

	OrderBookDisplay struct {
		Asks            []*Order
		Bids            []*Order
		TotalAsksVolume float64
		TotalBidsVolume float64
	}

	RegisterUserRequest struct {
		Username   string `json:"username"`
		Address    string `json:"address"`
		PrivateKey string `json:"privateKey"`
	}
)

const (
	MarketBTC   Market    = "BTC"
	MarketETH   Market    = "ETH"
	LimitOrder  OrderType = "LIMIT"
	MarketOrder OrderType = "MARKET"
)

func NewExchange(ethClient *ethclient.Client, hotWalletAddress, hotWalletPrivateKey string) (*Exchange, error) {
	privateKey, err := crypto.HexToECDSA(hotWalletPrivateKey)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &Exchange{
		userIdMapUser:    make(map[int64]*user.User),
		orderIdMapUserId: make(map[int64]int64),
		orderbooks:       make(map[Market]*orderbook.OrderBook),
		address:          hotWalletAddress,
		privateKey:       privateKey,
		ethClient:        ethClient,
	}, nil
}

func (ex *Exchange) InitOrderbooks() {
	ex.orderbooks[MarketBTC] = orderbook.NewOrderBook("BTC")
	ex.orderbooks[MarketETH] = orderbook.NewOrderBook("ETH")
}
func (ex *Exchange) HandlePlaceOrder(c echo.Context) error {

	var placeOrderData PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return c.JSON(400, map[string]any{"msg": "invalid request body"})
	}

	if placeOrderData.UserID == 0 {
		return c.JSON(400, map[string]any{"msg": "userId is required"})
	}

	market := Market(placeOrderData.Market)
	orderType := OrderType(placeOrderData.Type)
	ob := ex.orderbooks[market]

	if ob == nil {
		return c.JSON(400, "{'error': 'invalid market'}")
	}

	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)

	switch orderType {
	case LimitOrder:
		err := ex.handlePlaceLimitOrder(ob, order, placeOrderData.Price)
		if err != nil {
			log.Println(err)
			return c.JSON(http.StatusBadRequest, map[string]any{"msg": "failed to place limit order"})
		}
		return c.JSON(http.StatusOK, map[string]any{"msg": "limit order placed"})
	case MarketOrder:
		matches, respMatchOrders := ex.handlePlaceMarketOrder(ob, order, placeOrderData.Price)

		// handle matches (token transfer)
		if err := ex.handleMatches(matches); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]any{"matches": respMatchOrders})
	default:
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "missing order type"})
	}
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

	// TODO: should ensure request sender is order owner (secure)

	market := Market(marketStr)
	orderId, _ := strconv.Atoi(orderIdStr)
	ob := ex.orderbooks[market]
	order := ob.GetOrderById(int64(orderId))
	ob.CancelOrder(order)

	symbol := ob.Symbol
	userId := order.UserID
	user := ex.userIdMapUser[userId]
	fmt.Println("[HandleDeleteOrder] refund to user:", user.Username)
	fmt.Println("[HandleDeleteOrder] refund to address:", user.Address)
	fmt.Println("[HandleDeleteOrder] refund size:", order.Size)
	chainUtil.TransferToken(ex.ethClient, symbol, order.Size, user.Address, *ex.privateKey)

	return c.JSON(http.StatusOK, map[string]any{"msg": "OK"})
}

func (ex *Exchange) HandleGetOrderIds(c echo.Context) error {
	marketStr := c.Param("market")
	market := Market(marketStr)
	ob := ex.orderbooks[market]

	return c.JSON(http.StatusOK, map[string]any{"msg": ob.GetLimitOrderIds()})

}

func (ex *Exchange) handlePlaceLimitOrder(ob *orderbook.OrderBook, order *orderbook.Order, price float64) error {
	user := ex.userIdMapUser[order.UserID]
	if user == nil {
		return errors.New("user not found by ID")
	}

	// check user balance logic here.
	ex.checkUserBalance(user, order, price, ob.Symbol)
	amount := order.Size
	ex.transferToken(ob.Symbol, amount, ex.address, *user.PrivateKey)

	if price != 0 {
		ob.PlaceLimitOrder(price, order)
		return nil
	} else {
		// raise error
		return errors.New("limit order price can not be zero")
	}
}

func (ex *Exchange) handlePlaceMarketOrder(ob *orderbook.OrderBook, order *orderbook.Order, price float64) ([]orderbook.Match, []*Order) {
	matches := ob.PlaceMarketOrder(order)

	respMatchOrders := make([]*Order, 0, len(matches))
	for _, match := range matches {
		// extract match order bid or ask (if the order is a bid, the match order is an ask)
		isBid := !order.Bid
		// extract match order's ID
		var matchOrderId int64 = 0
		if isBid {
			matchOrderId = match.Bid.ID
		} else {
			matchOrderId = match.Ask.ID
		}

		respMatchOrders = append(respMatchOrders, &Order{
			Price:     match.Price,
			Size:      match.SizeFilled,
			Timestamp: time.Now().UnixNano(),
			Bid:       isBid,
			ID:        matchOrderId,
		})
	}
	return matches, respMatchOrders
}

func (ex *Exchange) handleMatches(matches []orderbook.Match) error {
	return nil
}

func (ex *Exchange) checkUserBalance(user *user.User, order *orderbook.Order, price float64, tokenSymbol string) bool {
	if order.Bid {
		// buy order should check user's USDT balance price*size
		amount := order.Size * price
		log.Println("[checkUserBalance] check user", user.Username, "USDT balance, should greater than ", amount)
		// check user USDT balance
		chainUtil.CheckBalance(ex.ethClient, "USDT", user.Address, amount)
		return true
	} else {
		// sell order should check user's ERC20 token balance
		log.Println("[checkUserBalance] check user ", user.Username, ", ", tokenSymbol, " balance, should greater than ", order.Size)
		// check user ERC20 Token size (symbol & size)
		chainUtil.CheckBalance(ex.ethClient, tokenSymbol, user.Address, order.Size)
		return true
	}
}

func (ex *Exchange) RegisterUser(c echo.Context) error {
	var registerUserData RegisterUserRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&registerUserData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "invalid request body"})
	}

	if registerUserData.Address == "" || registerUserData.Username == "" || registerUserData.PrivateKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "missing params"})
	}

	newUser := user.NewUser(registerUserData.Username, registerUserData.Address, registerUserData.PrivateKey)
	ex.userIdMapUser[newUser.UserID] = newUser
	return c.JSON(http.StatusOK, map[string]any{"userId": newUser.UserID})
}

func (ex *Exchange) transferToken(symbol string, amount float64, to string, privateKey ecdsa.PrivateKey) {
	chainUtil.TransferToken(ex.ethClient, symbol, amount, to, privateKey)
}

func (ex *Exchange) QueryBalance(c echo.Context) error {
	userIdStr := c.Param("userId")
	symbol := c.Param("symbol")

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "userId format incorrect"})
	}

	user, ok := ex.userIdMapUser[userId]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "user not found by ID"})
	}

	balance, err := ex.queryBalance(user, symbol)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "failed to query balance"})
	}
	return c.JSON(http.StatusOK, map[string]any{"symbol": symbol, "balance": balance})
}

func (ex *Exchange) queryBalance(user *user.User, symbol string) (float64, error) {
	return chainUtil.QueryBalance(ex.ethClient, symbol, user.Address)
}
