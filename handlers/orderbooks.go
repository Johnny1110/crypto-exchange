package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/labstack/gommon/log"
	"net/http"
)

func OrderbooksSnapshot(c *gin.Context) {
	market := c.Param("market")
	engine := c.MustGet("engine").(*core.MatchingEngine)

	ob, err := engine.GetOrderBook(market)
	if err != nil {
		log.Warn("Error getting orderbook: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	snapshot := ob.Snapshot()
	c.JSON(http.StatusOK, gin.H{"data": snapshot})
}
