# OrderBooks API

<br>

## OrderBook Snapshot

Display top 20 bids/asks price volume pair

URI: `/api/v1/orderbooks/{market}/snapshot`

Method: GET

Path-Param:
```
market: string (e.g. ETH-USDT, BTC-USDT, DOT-USDT)
```

<br>

Response-Body:

```json
{
    "code": "0000000",
    "message": "success",
    "data": {
        "BidSide": [
            {
                "price": 2700,
                "volume": 10
            },
            {
                "price": 2650,
                "volume": 10
            },
            {
                "price": 2600,
                "volume": 10
            },
            {
                "price": 2550,
                "volume": 10
            },
            {
                "price": 2500,
                "volume": 11
            }
        ],
        "AskSide": [
            {
                "price": 2500,
                "volume": 10
            },
            {
                "price": 2550,
                "volume": 10
            },
            {
                "price": 2600,
                "volume": 11
            },
            {
                "price": 2650,
                "volume": 10
            },
            {
                "price": 2700,
                "volume": 10
            }
        ]
    }
}
```