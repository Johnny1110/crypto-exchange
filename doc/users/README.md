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
    "data": {
        "userId": "92b229e2-447c-4775-8ad8-320b0b492e3d"
    },
    "message": "success"
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

```
{
    "code": "0000000",
    "data": {
        "token": "5e5c2694-9b3b-4097-8381-f36a55745117"
    },
    "message": "success"
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
    "data": null,
    "message": "success"
}
```