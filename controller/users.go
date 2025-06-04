package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/service"
	"net/http"
)

type UserController struct {
	userService service.IUserService
}

func (c UserController) Register(context *gin.Context) {
	var req dto.RegisterReq
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, HandleInvalidInput())
		return
	}

	userId, err := c.userService.Register(context.Request.Context(), &req)
	if err != nil {
		context.JSON(http.StatusBadRequest, HandleCodeError(REGISTER_ERROR, err))
		return
	}

	context.JSON(http.StatusOK, HandleSuccess(map[string]any{"user_id": userId}))
	return
}

func (c UserController) Login(context *gin.Context) {
	var req dto.LoginReq
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, HandleInvalidInput())
		return
	}
	token, err := c.userService.Login(context.Request.Context(), &req)
	if err != nil {
		context.JSON(http.StatusBadRequest, HandleCodeError(LOGIN_ERROR, err))
		return
	}
	context.JSON(http.StatusOK, HandleSuccess(map[string]any{"token": token}))
	return
}

func (c UserController) GetProfile(context *gin.Context) {
	userId := context.MustGet("userId").(string)
	user, err := c.userService.GetUser(context.Request.Context(), userId)
	if err != nil {
		context.JSON(http.StatusBadRequest, HandleCodeError(USER_DATA_NOT_FOUND, err))
		return
	}
	context.JSON(http.StatusOK, HandleSuccess(user))
	return
}

func (c UserController) Logout(context *gin.Context) {
	token := context.MustGet("token").(string)
	err := c.userService.Logout(context.Request.Context(), token)
	if err != nil {
		context.JSON(http.StatusBadRequest, HandleCodeError(USER_DATA_NOT_FOUND, err))
		return
	}
	context.JSON(http.StatusOK, HandleSuccess(nil))
	return
}

func NewUserController(userService service.IUserService) *UserController {
	return &UserController{
		userService: userService,
	}
}
