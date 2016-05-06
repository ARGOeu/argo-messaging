---
title: 'ARGO Messaging Service documentation | ARGO'
page_title: Argo Messaging API Authentication
font_title: fa fa-cogs
description: Argo Messaging API Information on authentication
---


# Authentication

Each user is authenticated by adding the url parameter ?key=T0K3N in each API request

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
