package controller

import (
	"time"
)

type Resp struct {
	Code      MessageCode `json:"code"`
	Msg       string      `json:"message"`
	Timestamp int64       `json:"timestamp"`
	Data      any         `json:"data"`
}

func HandleError(err error) any {
	return &Resp{
		Code:      SYSTEM_ERROR,
		Msg:       err.Error(),
		Timestamp: time.Now().UnixMilli(),
	}
}

func HandleCodeError(code MessageCode, err error) any {
	return &Resp{
		Code:      code,
		Msg:       err.Error(),
		Timestamp: time.Now().UnixMilli(),
	}
}

func HandleCodeErrorAndMsg(code MessageCode, msg string) any {
	return &Resp{
		Code:      code,
		Msg:       msg,
		Timestamp: time.Now().UnixMilli(),
	}
}

func HandleSuccess(data any) any {
	return &Resp{
		Code:      SUCCESS,
		Msg:       "success",
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
}

func HandleInvalidInput() any {
	return HandleCodeErrorAndMsg(INVALID_PARAMS, "invalid input")
}

type MessageCode string

const (
	SUCCESS MessageCode = "0000000"

	// common error: 1000001 ~ 1999999
	INVALID_PARAMS     = "1000001"
	FUNC_NOT_IMPLEMENT = "1000009"

	// users : 2000000 ~ 2999999
	REGISTER_ERROR      = "2000001"
	LOGIN_ERROR         = "2000002"
	USER_DATA_NOT_FOUND = "2000003"

	// orders : 3000000 ~ 3999999
	PLACE_ORDER_ERROR  = "3000001"
	CANCEL_ORDER_ERROR = "3000002"

	// balances : 4000000 ~ 4999999
	QUERY_BALANCE_ERROR = "4000001"

	// orderBooks: 5000000 ~ 5999999
	SNAPSHOT_ERROR = "5000001"

	BAD_REQUEST   MessageCode = "9000001"
	ACCESS_DENIED MessageCode = "9900001"
	SYSTEM_ERROR  MessageCode = "9999999"
)
