---
id: api_auth
title: Authentication
sidebar_position: 1
---

Each user is authenticated by adding the header parameter `x-api-key` in each API request

If a user does not provide a valid token the following response is returned:

```json
{
  "error": {
    "code": 401,
    "message": "Unauthenticated",
    "status": "UNAUTHENTICATED"
  }
}
```

The ARGO Messaging Service supports authorization. If a user is unauthorized the following response is returned:

```json
{
  "error": {
    "code": 403,
    "message": "Access to this resource is forbidden",
    "status": "FORBIDDEN"
  }
}
```
