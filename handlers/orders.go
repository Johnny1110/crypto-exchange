package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/engine-v2/book"
	"net/http"
)

// PlaceOrderHandler order entry
func PlaceOrderHandler(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	ob := c.MustGet("ob").(*book.OrderBook)

	userID := c.MustGet("userID").(string)
	market := c.Param("market") // router is /:market/order

	var req orderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: need find out which asset need to be frozen (by market & side)
	// ex ETH/USDT bid order -> freeze USDT
	// ex ETH/USDT ask order -> freeze ETH
	err := tryFreezeAmount(market, req.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "insufficient balance"})
	}

}

func tryFreezeAmount(market string, size float64) error {
	// TODO
	return nil
}
