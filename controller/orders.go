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
	user := context.MustGet("user").(*dto.User)
	market := context.Param("market") // router is /:market/order

	if user == nil || market == "" {
		context.JSON(http.StatusBadRequest, HandleInvalidInput())
		return
	}

	var req dto.OrderReq
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, HandleInvalidInput())
		return
	}

	log.Infof("[OrderContrller] Placing order: market:[%s], user:[%s], req: %v", market, user.Username, req)

	result, err := c.orderService.PlaceOrder(context.Request.Context(), market, user, &req)
	if err != nil {
		context.JSON(http.StatusBadRequest, HandleCodeError(PLACE_ORDER_ERROR, err))
		return
	}

	context.JSON(http.StatusOK, HandleSuccess(result))
}

func (c OrderController) CancelOrder(context *gin.Context) {
	userID := context.MustGet("userId").(string)
	orderID := context.Param("orderId")

	if userID == "" || orderID == "" {
		context.JSON(http.StatusBadRequest, HandleInvalidInput())
		return
	}

	log.Infof("[OrderController] Canceling order: userID:[%s], orderID: [%s]", userID, orderID)

	order, err := c.orderService.CancelOrder(context.Request.Context(), userID, orderID)
	if err != nil {
		context.JSON(http.StatusBadRequest, HandleCodeError(CANCEL_ORDER_ERROR, err))
		return
	}

	context.JSON(http.StatusOK, HandleSuccess(order))
}

func (c OrderController) GetOrders(ctx *gin.Context) {
	userID := ctx.MustGet("userId").(string)
	var query dto.GetOrdersQueryReq
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, HandleInvalidInput())
		return
	}

	query.UserID = userID

	resp, err := c.orderService.PaginationQuery(ctx.Request.Context(), &query)
	if err != nil {
		log.Errorf("[OrderController] failed to PaginationQuery, error: %v", err)
		ctx.JSON(http.StatusBadRequest, HandleInvalidInput())
		return
	}

	ctx.JSON(http.StatusOK, HandleSuccess(resp))
}
