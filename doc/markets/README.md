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