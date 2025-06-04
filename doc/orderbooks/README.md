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
    "timestamp": 1749024755669,
    "data": {
        "BidSide": [
            {
                "price": 3000,
                "volume": 10
            },
            {
                "price": 2900,
                "volume": 10
            },
            {
                "price": 2800,
                "volume": 10
            },
            {
                "price": 2700,
                "volume": 10
            },
            {
                "price": 2600,
                "volume": 10
            }
        ],
        "AskSide": [
            {
                "price": 3100,
                "volume": 9.9
            },
            {
                "price": 3200,
                "volume": 10
            },
            {
                "price": 3300,
                "volume": 10
            },
            {
                "price": 3400,
                "volume": 10
            },
            {
                "price": 3500,
                "volume": 10
            }
        ],
        "LatestPrice": 3100
    }
}
```