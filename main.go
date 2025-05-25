package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/johnny1110/crypto-exchange/engine-v1/exchange"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	hotWalletAddress    = "0x6812A4DB987C1bdaf14be68F8acD9c17948622b9"
	hotWalletPrivateKey = "b36b2f4b16ffbbbdb0369b597ce12ed24aaa729aa193cb12b6b070f4db566c7c"
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
