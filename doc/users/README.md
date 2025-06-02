# Users API

<br>

## Register

URI: `/users/register`
Method: POST
Request-Body:

```json
{
"username": "johnny",
"password": "1234"
}
```

<br>

## Login

URI: `/users/login`
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
"token": "94a2cc50-5478-48be-8cd5-d4fc486fa99c"
}
```

<br>

## Logout

URI: `localhost:8080/users/logout`
Method: DELETE

Header:

```
Authorization: string (login token)
```