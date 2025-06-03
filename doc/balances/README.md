# Balances API

<br>

## Get Balance Info

URI: `/api/v1/balances`

Method: GET

Header:
```
Authorization: string (login token)
```

Response-Body:

```json
{
    "code": "0000000",
    "data": [
        {
            "asset": "ASTR",
            "available": 0,
            "locked": 0,
            "total": 0
        },
        {
            "asset": "BTC",
            "available": 0,
            "locked": 0,
            "total": 0
        },
        {
            "asset": "DOT",
            "available": 0,
            "locked": 0,
            "total": 0
        },
        {
            "asset": "ETH",
            "available": 5,
            "locked": 0,
            "total": 5
        },
        {
            "asset": "HDX",
            "available": 0,
            "locked": 0,
            "total": 0
        },
        {
            "asset": "USDT",
            "available": 9000,
            "locked": 0,
            "total": 9000
        }
    ],
    "message": "success"
}
```

* `asset`: Currency
* `available`: Available amount
* `locked`: Means user have open order in orderbooks or asset is pending withdraw or pending deposit.

<br>
<br>

