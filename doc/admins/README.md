# Admins API (No need auth token just for now)

<br>

## Settlement Balance

URI: `/admin/api/v1/manual-adjustment`

Method: POST

Headers:

```
Admin-Token: string (using 'frizo' for testing)
```

Request-Body:
```json
{
    "username": "johnny",
    "amount": 3000,
    "asset": "USDT"
}
```

Response-Body:

```json
{
    "code": "0000000",
    "message": "success",
    "timestamp": 1749024602941,
    "data": null
}
```

<br>