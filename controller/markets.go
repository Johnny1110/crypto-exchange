package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/service"
	"net/http"
)

type MarketDataController struct {
	marketDataService service.IMarketDataService
}

func NewMarketDataController(marketDataService service.IMarketDataService) *MarketDataController {
	return &MarketDataController{marketDataService: marketDataService}
}

func (mc MarketDataController) GetAllMarketsData(ctx *gin.Context) {
	data, err := mc.marketDataService.GetAllMarketData()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, HandleError(err))
		return
	}

	var markets []interface{}
	for _, marketData := range data {
		markets = append(markets, marketData)
	}

	ctx.JSON(http.StatusOK, HandleSuccess(markets))
}

func (mc MarketDataController) GetMarketsData(ctx *gin.Context) {
	market := ctx.Param("market")
	if market == "" {
		ctx.JSON(http.StatusBadRequest, HandleInvalidInput())
		return
	}
	data, err := mc.marketDataService.GetMarketData(market)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err))
		return
	}

	ctx.JSON(http.StatusOK, HandleSuccess(data))
}
