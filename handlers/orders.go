package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/service"
	"net/http"
)

// PlaceOrderHandler order entry
func PlaceOrderHandler(c *gin.Context) {
	engine := c.MustGet("engine").(*core.MatchingEngine)
	orderService := c.MustGet("orderService").(*service.OrderService)
	userID := c.MustGet("userID").(string)

	market := c.Param("market") // router is /:market/order

	var err error
	if !engine.ValidateMarket(market) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Market"})
		return
	}

	var req orderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	result, err := orderService.PlaceOrder(service.PlaceOrderRequest{
		UserID:    userID,
		Market:    market,
		Side:      req.Side,
		Price:     req.Price,
		Size:      req.Size,
		OrderType: req.OrderType,
		Mode:      req.Mode,
	})

	if err != nil {
		if err.Error() == "insufficient balance" {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id": result.OrderID,
		"status":   result.Status,
		"trades":   result.Trades,
	})
}

func CancelOrderHandler(c *gin.Context) {
	engine := c.MustGet("engine").(*core.MatchingEngine)
	orderService := c.MustGet("orderService").(*service.OrderService)
	userID := c.MustGet("userID").(string)
	orderID := c.Param("orderID")

	market := c.Param("market") // router is /:market/order

	var err error
	if !engine.ValidateMarket(market) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Market"})
		return
	}

	err = orderService.CancelOrder(market, userID, orderID)
	if err != nil {
		if err.Error() == "Order Not Exist" {
			c.JSON(http.StatusNotFound, gin.H{"message": "Order Not Exist"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order Canceled"})
}
