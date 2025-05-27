package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/service"
)

func AutoMarketMaker(c *gin.Context) {
	autoMakerService := c.MustGet("autoMakerService").(*service.AutoMakerService)
	orderService := c.MustGet("orderService").(*service.OrderService)

	// TODO: refactor this
	autoMakerService.MakeMarket(orderService)
}
