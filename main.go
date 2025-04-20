package main

import (
	"github.com/johnny1110/crypto-exchange/exchange"
	"github.com/labstack/echo/v4"
	"net/http"
)

func main() {
	e := echo.New()

	ex := exchange.NewExchange()
	ex.InitOrderbooks()
	e.GET("/healthcheck", func(c echo.Context) error { return c.String(http.StatusOK, "OK") })
	e.POST("/order", ex.HandlePlaceOrder)
	e.DELETE("/orderbook/:market/order/:id", ex.HandleDeleteOrder)
	e.GET("/orderbook/:market", ex.HandleGetOrderBook)
	e.GET("/orderbook/:market/orderIds", ex.HandleGetOrderIds)

	e.Start(":3000")
}
