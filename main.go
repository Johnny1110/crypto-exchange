package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/johnny1110/crypto-exchange/exchange"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	ex := exchange.NewExchange()
	ex.InitOrderbooks()
	e.GET("/healthcheck", func(c echo.Context) error { return c.String(http.StatusOK, "OK") })
	e.POST("/order", ex.HandlePlaceOrder)
	e.DELETE("/orderbook/:market/order/:id", ex.HandleDeleteOrder)
	e.GET("/orderbook/:market", ex.HandleGetOrderBook)
	e.GET("/orderbook/:market/orderIds", ex.HandleGetOrderIds)

	// testing ganache with a hot wallet balance check.
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal("Failed to connect to the Ethereum client:", err)
	}

	ctx := context.Background()
	address := common.HexToAddress("0xbD89F049A4849235cff2B6Ca08b53Bee079AB277")
	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		log.Fatal("Failed to get balance:", err)
	}

	fmt.Println("Address:", address.Hex())
	fmt.Println("Balance:", balance)

	e.Start(":3000")
}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}
