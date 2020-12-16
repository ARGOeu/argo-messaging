# Introduction

Each user is authenticated by adding the url parameter `?key=T0K3N` in each API request

Users can also authenticate using the header `x-api-key`.

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
