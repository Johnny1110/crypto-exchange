# Document

<br>

## Module Category

* healthcheck:

```
URI: /health
Method: GET
```

<br>

## Currently Supported Markets:

```
BTC-USDT
ETH-USDT
DOT-USDT
```

<br>
<br>

## API Response Format

```json
{
    "code": string,
    "message": string,
    "timestamp": number (unix millsec),
    "data": object
}
```

<br>
<br>

## API 

* [Users](users)
* [Balances](balances)
* [Orders](orders)
* [OrderBooks](orderbooks)
* [Market][markets]

* [Admins (General for testing)](admins)

<br>
<br>

## API Message Code

message code
```
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

```
