package handlers

import "github.com/gin-gonic/gin"

func HandleError(err error) map[string]any {
	return gin.H{"code": SYSTEM_ERROR, "message": err.Error()}
}

func HandleCodeError(code MessageCode, err error) map[string]any {
	return gin.H{"code": code, "message": err.Error()}
}

func HandleSuccess(data any) map[string]any {
	return gin.H{"code": 0000000, "message": SUCCESS, "data": data}
}

type MessageCode string

const (
	SUCCESS MessageCode = "0000000"

	// user : 1000000 ~ 1999999
	REGISTER_ORDER_ERROR = "1000001"
	LOGIN_ERROR          = "1000002"

	// order : 2000000 ~ 2999999
	PLACE_ORDER_ERROR = "2000001"

	BAD_REQUEST   MessageCode = "900001"
	ACCESS_DENIED MessageCode = "990001"
	SYSTEM_ERROR  MessageCode = "9999999"
)
