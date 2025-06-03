package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/service"
	"net/http"
)

type BalanceController struct {
	balanceService service.IBalanceService
}

func (c BalanceController) GetBalances(context *gin.Context) {
	userId := context.MustGet("userId").(string)
	balances, err := c.balanceService.GetBalances(context.Request.Context(), userId)
	if err != nil {
		context.JSON(http.StatusBadRequest, HandleCodeError(QUERY_BALANCE_ERROR, err))
		return
	}
	context.JSON(http.StatusOK, HandleSuccess(balances))
}

func NewBalanceController(balanceService service.IBalanceService) *BalanceController {
	return &BalanceController{
		balanceService: balanceService,
	}
}
