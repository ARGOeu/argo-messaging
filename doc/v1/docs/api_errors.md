# Errors

In case of Error during handling user’s request the API responds using the following schema:

```json
{
   "error": {
      "code": 500,
      "message": "Something bad happened",
      "status": "INTERNAL"
   }
}
```
## Captured Errors from usage scenarios

Error | Code | Response | Related Requests
------|------|----------|------------------
Topic already exists | 409 | ```{"error":{"code":409,"message":"Topic already exists","status":"ALREADY_EXISTS"}}``` | Create Topic (PUT)  
Subscription already exists | 409 | ```{"error":{"code":409,"message":"Subscription already exists","status":"ALREADY_EXISTS"}}``` | Create Subscription (PUT)
Invalid Topics Name | 400 | ```{"error":{"code":400,"message":"Invalid topics name","status":"INVALID_ARGUMENT"}}``` | Create Subscription (PUT)
Topic Doesn't Exist | 404 | ```{"error":{"code":404,"message":"Topic does not exist","status":"NOT_FOUND"}}``` | Show specific Topic  (GET)
Invalid Topic ACL arguments | 400 | ```{"error":{"code":400,"message":"Invalid Topic ACL Arguments","status":"INVALID_ARGUMENT"}}``` | Modify Topic ACL (POST)
Subscription Doesn't Exist | 404 | ```{"error":{"code":404,"message":"Subscription does not exist","status":"NOT_FOUND"}}``` | Show specific Subscription  (GET)
Message size to large | 413 | ```{"error":{"code":413,"message":"Message size too large","status":"INVALID_ARGUMENT"}}``` | Topic Publish (POST)
Invalid Subscription Arguments | 400 | ```{"error":{"code":404,"message":"Invalid Subscription Arguments","status":"INVALID_ARGUMENT"}}``` | Create Subscription (POST), Modify Push Configuration (POST)
Invalid Subscription ACL arguments | 400 | ```{"error":{"code":400,"message":"Invalid Subscription ACL Arguments","status":"INVALID_ARGUMENT"}}``` | Modify Subscription ACL (POST)
Invalid ACK Parameter | 400 | ```{"error":{"code":400,"message":"Invalid ack parameter","status":"INVALID_ARGUMENT"}}``` | Subscription Acknowledge (POST)
Invalid ACK id | 400 | ```{"error":{"code":400,"message":"Invalid ack id parameter","status":"INVALID_ARGUMENT"}}``` | Subscription Acknowledge (POST)
Invalid pull parameters | 400 | ```{"error":{"code":400,"message":"Pull Parameters Invalid","status":"INVALID_ARGUMENT"}}``` | Subscription Pull (POST)
Unauthorized | 401 | ```{"error":{"code":401,"message":"Unauthorized","status":"UNAUTHORIZED"}}``` | All requests _(if a user is not authenticated)_
Forbidden Access to Resource  | 403 | ```{"error":{"code":403,"message":"Access to this resource is forbidden","status":"FORBIDDEN"}}``` | All requests _(if a user is forbidden to access the resource)_
