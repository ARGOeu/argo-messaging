# Introduction

Each user is authenticated by using the header `x-api-key`.

## Εxample 

For example the status of the service can be seen by making the following request

```
curl -X GET -H “Content-Type: application/json”  -H "x-api-key:S3CR3T"  “https://{URL}/v1/status
```

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
