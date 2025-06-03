package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/service"
	"net/http"
)

type AdminController struct {
	adminService service.IAdminService
}

func (c AdminController) ManualAdjustment(context *gin.Context) {
	var req dto.SettlementReq
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, HandleCodeErrorAndMsg(INVALID_PARAMS, "input parameter error"))
		return
	}

	err := c.adminService.Settlement(context.Request.Context(), req)
	if err != nil {
		context.JSON(http.StatusBadRequest, HandleCodeError(INVALID_PARAMS, err))
		return
	}

	context.JSON(http.StatusOK, HandleSuccess(nil))
}

func (c AdminController) TestMakeMarket(context *gin.Context) {
	// TODO: implement auto market maker logic.
	context.JSON(http.StatusBadRequest, HandleCodeErrorAndMsg(FUNC_NOT_IMPLEMENT, "func not support yet"))
	return
}

func NewAdminController(adminService service.IAdminService) *AdminController {
	return &AdminController{
		adminService: adminService,
	}
}
