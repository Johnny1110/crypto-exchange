# Websocket

<br>

## Endpoint

* local: `localhost:8081/ws`
* remote: `34.81.155.101:8081/ws`

<br>

## Message Form

```json
{
  "action": string, // mandatory
  "channel": string, // mandatory
  "params": object // optional
}
```

* Supported Action
    * `subscribe` subscribe 1 channel with param
    * `unsubscribe` unsubscribe 1 channel with param

<br>

* Supported Channel
    * `ohlcv` provide realtime symbol ohlcv data with interval.
    * `orderbook` provide target market orderbook data.
    * `markets`  provide all markets data.

<br>

## Channel

<br>

* OHLCV

  subscribe message:

    ```json
    {
        "action": "subscribe",
        "channel": "ohlcv",
        "params": {
            "symbol": "ETH-USDT",
            "interval": "15m"
        }
    }
    ```

    unsubscribe message:
    ```json
    {
        "action": "unsubscribe",
        "channel": "ohlcv",
        "params": {
            "symbol": "ETH-USDT",
            "interval": "15m"
        }
    }
    ```
  
    support interval: `15m`, `1h`, `1d`, `1w`.

    <br>

    response:
    
    ```json
    {
      "channel": "ohlcv",
      "data": {
        "s": "ok",
        "t": [
          1750440600
        ],
        "o": [
          0.001
        ],
        "h": [
          0.001
        ],
        "l": [
          0.001
        ],
        "c": [
          0.001
        ],
        "v": [
          0
        ]
      },
      "timestamp": 1750441310
    }
    ```
  

<br>

* OrderBook

  subscribe message:

    ```json
    {
        "action": "subscribe",
        "channel": "orderbook",
        "params": {
            "market": "ETH-USDT",
        }
    }
    ```

  unsubscribe message:
    ```json
    {
        "action": "unsubscribe",
        "channel": "orderbook",
        "params": {
            "market": "ETH-USDT",
        }
    }
    ```
  
   response:

    ```json
    {
        "channel": "orderbook",
        "data": {
            "bid_side": [
                {
                "price": 2560.6,
                "volume": 112.5268843
                },
              ...
            ],
            "ask_side": [
                {
                "price": 2485.38,
                "volume": 1.609
                },
              ...
          ],
            "latest_price": 2490.89,
            "best_bid_price": 2560.6,
            "best_ask_price": 2485.38,
            "total_bid_size": 145.09088429999994,
            "total_ask_size": 32.655230800000005
            },
        "timestamp": 1750438357
    }
    ```

<br>

* All Markets

  subscribe message:

    ```json
    {
        "action": "subscribe",
        "channel": "markets"
    }
    ```

  unsubscribe message:
    ```json
    {
        "action": "unsubscribe",
        "channel": "markets"
    }
    ```
  
    response:
    ```json
    {
        "channel": "markets",
        "data": [
            {
                "market_name": "BTC-USDT",
                "latest_price": 0,
                "price_change_24h": 0,
                "total_volume_24h": 0
            },
            {
                "market_name": "ETH-USDT",
                "latest_price": 2490.89,
                "price_change_24h": 0,
                "total_volume_24h": 8.508905207460849
            },
            {
                "market_name": "DOT-USDT",
                "latest_price": 0,
                "price_change_24h": 0,
                "total_volume_24h": 0
            },
            ...
        ],
        "timestamp": 1750438850
    } 
   ```