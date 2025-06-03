package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/service"
	"net/http"
)

type OrderBookController struct {
	service service.IOrderBookService
}

func (c OrderBookController) OrderbooksSnapshot(context *gin.Context) {
	market := context.Param("market")
	snapshot, err := c.service.GetSnapshot(context.Request.Context(), market)
	if err != nil {
		context.JSON(http.StatusBadRequest, HandleCodeError(SNAPSHOT_ERROR, err))
		return
	}

	context.JSON(http.StatusOK, HandleSuccess(snapshot))
}

func NewOrderBookController(service service.IOrderBookService) *OrderBookController {
	return &OrderBookController{
		service: service,
	}
}
