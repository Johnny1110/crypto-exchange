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
    "price": $2600,
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
"quote_amount": $300
}
```

<br>
<br>

## Query Order

TODO: query open order.
TODO: query dealt order.