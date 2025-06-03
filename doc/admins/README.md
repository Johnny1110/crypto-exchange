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
    "secret": "frizo", // static string
    "amount": 3000,
    "asset": "USDT"
}
```

<br>


## Trigger Auto Market Maker (For Testing)

URI: `/admin/api/v1/auto-market-maker`

Method: POST

Headers:

```
Admin-Token: string (using 'frizo' for testing)
```