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
    "timestamp": 1749182627291,
    "data": {
        "bid_side": [
            {
                "price": 3000,
                "volume": 9.100000000000001
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
        "ask_side": [
            {
                "price": 3100,
                "volume": 9.290322580645151
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
    "latest_price": 3000,
    "best_bid_price": 3000,
    "best_ask_price": 3100
    }
}
```
