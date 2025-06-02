# Balances API

<br>

## Get Balance Info

URI: `/balances`
Method: GET
Header:
```
Authorization: string (login token)
```

Response-Body:

```json
[
    {
        "asset": "ASTR",
        "available": 0,
        "locked": 0
    },
    {
        "asset": "BTC",
        "available": 0,
        "locked": 0
    },
    {
        "asset": "DOT",
        "available": 0,
        "locked": 0
    },
    {
        "asset": "ETH",
        "available": 5,
        "locked": 0
    },
    {
        "asset": "HDX",
        "available": 0,
        "locked": 0
    },
    {
        "asset": "USDT",
        "available": 3000,
        "locked": 0
    }
]
```

* `asset`: Currency
* `available`: Available amount
* `locked`: Means user have open order in orderbooks or asset is pending withdraw or pending deposit.

<br>
<br>

