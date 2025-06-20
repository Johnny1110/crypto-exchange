package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/ohlcv"
	"github.com/johnny1110/crypto-exchange/service"
	"net/http"
	"strconv"
	"time"
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

func (mc MarketDataController) GetOHLCVHistory(ctx *gin.Context) {
	market := ctx.Param("market")
	intervalStr := ctx.Param("interval")

	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")
	limitStr := ctx.Query("limit")

	if market == "" || intervalStr == "" {
		ctx.JSON(http.StatusBadRequest, HandleInvalidInput())
		return
	}

	var startTime int64
	if startTimeStr != "" {
		parsed, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, HandleInvalidInput())
			return
		}
		startTime = parsed
	} else {
		// default half years ago.
		startTime = time.Now().AddDate(0, -6, 0).Unix()
	}

	var endTime int64
	if endTimeStr != "" {
		parsed, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, HandleInvalidInput())
			return
		}
		endTime = parsed
	} else {
		// default now
		endTime = time.Now().Unix()
	}

	if startTime >= endTime {
		ctx.JSON(http.StatusBadRequest, HandleInvalidInput())
		return
	}
	// if limitStr not input, default is 500
	var limit int32 = 500
	if limitStr != "" {
		parsed, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, HandleInvalidInput())
			return
		}
		if parsed <= 0 {
			ctx.JSON(http.StatusBadRequest, HandleInvalidInput())
			return
		}
		if parsed > 1000 {
			ctx.JSON(http.StatusBadRequest, HandleInvalidInput())
			return
		}
		limit = int32(parsed)
	}

	interval := ohlcv.OHLCV_INTERVAL(intervalStr)
	if _, ok := ohlcv.SupportedIntervals[interval]; !ok {
		ctx.JSON(http.StatusBadRequest, HandleInvalidInput())
	}
	data, err := mc.marketDataService.GetOHLCVHistory(ctx.Request.Context(), &ohlcv.GetOhlcvDataReq{
		Symbol:    market,
		Interval:  interval,
		StartTime: time.Unix(startTime, 0),
		EndTime:   time.Unix(endTime, 0),
		Limit:     int(limit),
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, HandleError(err))
		return
	}

	ctx.JSON(http.StatusOK, HandleSuccess(data))
}
