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
    "message": "success",
    "timestamp": 1749024630775,
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
            "available": 0,
            "locked": 0,
            "total": 0
        },
        {
            "asset": "HDX",
            "available": 0,
            "locked": 0,
            "total": 0
        },
        {
            "asset": "USDT",
            "available": 3150,
            "locked": 0,
            "total": 3150
        }
    ]
}
```

* `asset`: Currency
* `available`: Available amount
* `locked`: Means user have open order in orderbooks or asset is pending withdraw or pending deposit.

<br>
<br>

