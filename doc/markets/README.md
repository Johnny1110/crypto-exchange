# Market API

<br>

---

<br>

## Get All Market Data

Display top 20 bids/asks price volume pair

URI: `/api/v1/markets`

Method: GET

<br>

Response-Body:

```json
{
  "code": "0000000",
  "message": "success",
  "timestamp": 1749226383432,
  "data": [
    {
      "market_name": "BTC-USDT",
      "latest_price": 13012.13,
      "price_change_24h": 0.1,
      "total_volume_24h": 21.312
    },
    {
      "market_name": "ETH-USDT",
      "latest_price": 2100,
      "price_change_24h": -0.032,
      "total_volume_24h": 1501.433
    },
    {
      "market_name": "DOT-USDT",
      "latest_price": 4.12,
      "price_change_24h": -0.01,
      "total_volume_24h": 9812
    }
  ]
}
```

## Get Market Data

Display top 20 bids/asks price volume pair

URI: `/api/v1/market/{market}`

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
  "timestamp": 1749226383432,
  "data": {
      "market_name": "BTC-USDT",
      "latest_price": 13012.13,
      "price_change_24h": 0.1,
      "total_volume_24h": 21.312
  }
}
```

## Get Market OHLCV History Data

OHLCV Data for TradingView

URI: `/api/v1/markets/{market}/ohlcv-history/{interval}`

Method: GET

Path-Param:
```
market: string (e.g. ETH-USDT, BTC-USDT, DOT-USDT)
interval: string (e.g. 15m, 1h, 1d, 1w)
```

<br>

Request-Param:

```
start_time: timestamp (optional default: half year ago)
end_time: timestamp (optional default: now)
limit: number (optional default: 500)
```

<br>

Response-Body:

```json
{
  "code": "0000000",
  "message": "success",
  "timestamp": 1750441473348,
  "data": {
    "s": "ok",
    "t": [
      1750392000
    ],
    "o": [
      0
    ],
    "h": [
      2600
    ],
    "l": [
      0
    ],
    "c": [
      2520.35
    ],
    "v": [
      0.46883140467364887
    ]
  }
}
```