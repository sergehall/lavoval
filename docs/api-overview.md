# API Overview

## Public endpoints

- `GET /healthz`
- `GET /readyz`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/skills`
- `GET /api/v1/skills/{skillID}`

## Authenticated endpoints

- `POST /api/v1/auth/logout`
- `GET /api/v1/me`
- `PATCH /api/v1/me/profile`

## Admin endpoints

- `GET /api/v1/admin/users`
- `GET /api/v1/admin/skills`
- `POST /api/v1/admin/skills`
- `PATCH /api/v1/admin/skills/{skillID}`
- `DELETE /api/v1/admin/skills/{skillID}`

## Response envelope

Successful responses return:

```json
{
  "data": { }
}
```

Error responses return:

```json
{
  "error": {
    "code": "validation_error",
    "message": "human readable message"
  }
}
```
