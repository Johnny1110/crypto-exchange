# Orders API


<br>
<br>

## Place Limit Order

<br>

URI: `/api/v1/orders/{market}`

Method: POST

Headers:
```
Authorization: string (login token)
```

Request-Body:
```
{
    "side": number,
    "order_type": number,
    "mode": number,
    "price": number,
    "size": number
}
```

<br>

Response-Body:
```json
{
    "code": "0000000",
    "message": "success",
    "timestamp": 1749024658921,
    "data": {
        "matches": [
            {
                "price": 3100, // dealt price
                "size": 0.1, // dealt qty
                "timestamp": 1749024658915 // dealt time (millsec)
            }
        ],
        "order": {
            "id": "3029de80-2767-44dd-89b5-653a539a96d5", // request orderId
            "user_id": "bc5eef14-de49-4804-992c-cac6ea0db6fb",
            "market": "ETH-USDT",
            "side": 0, // 0:buy order, 1: sell order
            "price": 3150, // limit price (market order don't have price -> null)
            "original_size": 0.1, // request qty
            "remaining_size": 0, // 0 if all filled
            "quote_amount": 310, // buy order: total cost / sell order: total paid
            "avg_dealt_price": 3100, // avg dealt price
            "type": 0, // 0: limit order, 1: market order
            "mode": 1, // 0: maker, 1: taker
            "fees": 0.0006451612903225806, // trading fees
            "fee_asset": "ETH", // paid fee asset
            "fee_rate": "0.2000%",
            "status": "FILLED",
            "created_at": 1749024658911,
            "updated_at": 1749024658911
        }
    }
}
```

<br>

Params:

* side: 0=Bid(Buy), 1=-Ask(Sell)
* order_type: 0=Limit Order, 1= Market Order
* mode: 0=Maker, 1=Taker(user)
* price: required when order_type=0 (limit)
* size: required when order_type=0 (limit)
* quote_amount: required when order_type=1 (market)

<br>
<br>

## Example

<br>

### I want to put a __buy ETH__ order into OrderBook as a market maker, price limit is $2500 USDT, qty is 10.

URI: `/api/v1/orders/ETH-USDT`

Method: POST

Headers:
```
"Authorization": "94a2cc50-5478-48be-8cd5-d4fc486fa99c"
```

Request-Body:
```json
{
    "side": 0, //buy
    "order_type": 0, // limit
    "mode": 0, // maker
    "price": 2500,
    "size": 10
}
```

<br>

### I want to sell ETH by limit order as a user, price limit is $2600 USDT, qty is 0.131.

URI: `/api/v1/orders/ETH-USDT`

Method: POST

Headers:
```
"Authorization": "94a2cc50-5478-48be-8cd5-d4fc486fa99c"
```

Request-Body:
```json
{
    "side": 1, // sell
    "order_type": 0, // limit
    "mode": 1, // taker (user)
    "price": 2600,
    "size": 0.131
}
```

<br>

### I want to buy ETH by market order as a user, I only want to cost total $300 USDT.

URI: `/api/v1/orders/ETH-USDT`

Method: POST

Headers:
```
"Authorization": "94a2cc50-5478-48be-8cd5-d4fc486fa99c"
```

Request-Body:
```json
{
"side": 0, // sell
"order_type": 1, // market
"quote_amount": 300
}
```

<br>
<br>


## Cancel Order

URI: `/api/v1/orders/{order_id}`

Method: DELETE

Headers:
```
Authorization: string (login token)
```

Params:
```
order_id: string (mandatory)
```

Response-Body:

```json
{
    "code": "0000000",
    "message": "success",
    "timestamp": 1749146832324,
    "data": {
        "id": "c654f54e-3872-4cd5-84b8-92a70d2bfd23",
        "market": "ETH-USDT",
        "side": 0,
        "original_size": 0.5,
        "remaining_size": 0.5,
        "quote_amount": 0,
        "avg_dealt_price": 0,
        "type": 0,
        "mode": 1,
        "status": "CANCELED", // CANCELED
        "fees": 0,
        "fee_asset": "ETH",
        "price": 3001,
        "fee_rate": "0.2000%",
        "created_at": 1749146780831,
        "updated_at": 1749146780831
    }
}
```

<br>
<br>

## Query Order

<br>

URI: `/api/v1/orders`

Method: GET

Headers:
```
Authorization: string (login token)
```

Params:
```
market: string (optional) e,g, "ETH-USDT"
side: number (optional) 0: buy-order 1: sell-order
type: string (mandatory) "OPENING", "CLOSED"
page_size: number (optioanl) default=10
current_page: number (optioanl) default=1
```

<br>

Response-Body:

```json
{
    "code": "0000000",
    "message": "success",
    "timestamp": 1749147236137,
    "data": {
        "total": 20,
        "total_page": 2,
        "current_page": 1,
        "page_size": 10,
        "has_next": true,
        "has_prev": false,
        "result": [
            {
                "id": "c654f54e-3872-4cd5-84b8-92a70d2bfd23",
                "market": "ETH-USDT",
                "side": 0,
                "original_size": 0.5,
                "remaining_size": 0.5,
                "quote_amount": 0,
                "avg_dealt_price": 0,
                "type": 0,
                "mode": 1,
                "status": "CANCELED",
                "fees": 0,
                "fee_asset": "ETH",
                "price": 3001,
                "fee_rate": "0.2000%",
                "created_at": 1749146780831,
                "updated_at": 1749146832314
            },
            {
                "id": "dd1853bf-1266-4f24-b386-7a4e3c47a9ce",
                "market": "ETH-USDT",
                "side": 1,
                "original_size": 0.15,
                "remaining_size": 0,
                "quote_amount": 450,
                "avg_dealt_price": 3000,
                "type": 1,
                "mode": 1,
                "status": "FILLED",
                "fees": 0.9,
                "fee_asset": "USDT",
                "fee_rate": "0.2000%",
                "created_at": 1749146639744,
                "updated_at": 1749146639754
            },
            ...
        ]
    }
}
```
