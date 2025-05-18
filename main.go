package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"net/http"

	"github.com/johnny1110/crypto-exchange/exchange"
	"github.com/labstack/echo/v4"
)

const (
	hotWalletAddress    = "0xD7e264213909b9b2D5e5164c1973191aAeb9e591"
	hotWalletPrivateKey = "97e64e8033170f2399b50bd57edf7629605206585dc59dcfc24de3c08dfd1d0c"
)

func main() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	ethClient, err := getEthClient()
	if err != nil {
		e.Logger.Fatal(err)
	}
	ex, err := exchange.NewExchange(ethClient, hotWalletAddress, hotWalletPrivateKey)

	if err != nil {
		log.Fatal(err)
	}

	ex.InitOrderbooks()
	e.GET("/healthcheck", func(c echo.Context) error { return c.String(http.StatusOK, "OK") })
	e.POST("/order", ex.HandlePlaceOrder)
	e.DELETE("/orderbook/:market/order/:id", ex.HandleDeleteOrder)
	e.GET("/orderbook/:market", ex.HandleGetOrderBook)
	e.GET("/orderbook/:market/orderIds", ex.HandleGetOrderIds)
	e.POST("/user/register", ex.RegisterUser)
	e.GET("/user/:userId/symbol/:symbol/balance", ex.QueryBalance)

	e.Start(":3000")
}

func getEthClient() (*ethclient.Client, error) {
	return ethclient.Dial("http://localhost:8545")
}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}
