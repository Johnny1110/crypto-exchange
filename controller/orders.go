package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/labstack/gommon/log"
	"net/http"
)

type OrderController struct {
	orderService service.IOrderService
}

func NewOrderController(orderService service.IOrderService) *OrderController {
	return &OrderController{
		orderService: orderService,
	}
}

func (c OrderController) PlaceOrder(context *gin.Context) {
	userID := context.MustGet("userId").(string)
	market := context.Param("market") // router is /:market/order

	if userID == "" || market == "" {
		context.JSON(http.StatusBadRequest, HandleInvalidInput())
		return
	}

	var req dto.OrderReq
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, HandleInvalidInput())
		return
	}

	log.Infof("[OrderContrller] Placing order: market:[%s], userID:[%s], req: %v", market, userID, req)

	result, err := c.orderService.PlaceOrder(context.Request.Context(), market, userID, &req)
	if err != nil {
		context.JSON(http.StatusBadRequest, HandleCodeError(PLACE_ORDER_ERROR, err))
		return
	}

	context.JSON(http.StatusOK, HandleSuccess(result))
}

func (c OrderController) CancelOrder(context *gin.Context) {
	userID := context.MustGet("userId").(string)
	orderID := context.Param("orderId")
	market := context.Param("market") // router is /:market/order

	if userID == "" || orderID == "" || market == "" {
		context.JSON(http.StatusBadRequest, HandleInvalidInput())
		return
	}

	log.Infof("[OrderController] Canceling order: market:[%s], userID:[%s], orderID: [%s]", market, userID, orderID)

	order, err := c.orderService.CancelOrder(context.Request.Context(), market, userID, orderID)
	if err != nil {
		context.JSON(http.StatusBadRequest, HandleCodeError(CANCEL_ORDER_ERROR, err))
		return
	}

	context.JSON(http.StatusOK, HandleSuccess(order))
}
