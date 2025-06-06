# Users API

<br>

## Register

URI: `/api/v1/users/register`

Method: POST

Request-Body:

```json
{
    "username": "johnny",
    "password": "1234"
}
```

Response-Body:

```json
{
    "code": "0000000",
    "message": "success",
    "timestamp": 1749025140955,
    "data": {
        "user_id": "7c775fbc-c907-4f26-8d33-05e6820875c7"
    }
}
```

<br>

## Login

URI: `/api/v1/users/login`

Method: POST

Request-Body:

```json
{
    "username": "johnny",
    "password": "1234"
}
```

Response-Body:

```json
{
    "code": "0000000",
    "message": "success",
    "timestamp": 1749025156135,
    "data": {
        "token": "de1160b7-6935-4c24-b230-826902f54d84"
    }
}
```

<br>

## Logout

URI: `/api/v1/users/logout`

Method: POST

Header:

```
Authorization: string (login token)
```

Response-Body:

```json
{
    "code": "0000000",
    "message": "success",
    "timestamp": 1749025184954,
    "data": null
}
```

<br>

## Get User Profile

URI: `/api/v1/users/profile`

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
  "timestamp": 1749227089633,
  "data": {
    "id": "U01_the_GOD",
    "username": "johnny",
    "vip_level": 1,
    "maker_fee": 0.001,
    "taker_fee": 0.002,
    "created_at": 1749226781000
  }
}
```

* maker_fee: user's maker trading fee rate.

* taker_fee: user's taker trading fee rate.