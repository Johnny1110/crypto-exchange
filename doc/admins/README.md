# Admins API (No need auth token just for now)

<br>

## Settlement Balance

URI: `/admin/manual-adjustment`
Method: POST
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

URI: `/admin/auto-market-maker`
Method: POST