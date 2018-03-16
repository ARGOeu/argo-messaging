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
## Error Codes

The following error codes are the possinble errors of all methods

Error | Code | Status | Related Requests
------|------|----------|------------------
Ack Timeout | 408 | TIMEOUT | Acknowledge Message (POST) - [more info](overview.md#Message-acknowledgement-deadline)
Topic already exists | 409 | ALREADY_EXISTS | Create Topic (PUT)  
Subscription already exists | 409 | ALREADY_EXISTS | Create Subscription (PUT)
Invalid Topics Name | 400 | INVALID_ARGUMENT | Create Subscription (PUT)
Topic Doesn't Exist | 404 | NOT_FOUND | Show specific Topic  (GET)
Invalid Topic ACL arguments | 400 | INVALID_ARGUMENT | Modify Topic ACL (POST)
Subscription Doesn't Exist | 404 | NOT_FOUND | Show specific Subscription  (GET)
Message size to large | 413 | INVALID_ARGUMENT | Topic Publish (POST)
Invalid Subscription Arguments | 400 | INVALID_ARGUMENT | Create Subscription (POST), Modify Push Configuration (POST)
Invalid Subscription ACL arguments | 400 | INVALID_ARGUMENT | Modify Subscription ACL (POST)
Invalid ACK Parameter | 400 | INVALID_ARGUMENT | Subscription Acknowledge (POST)
Invalid ACK id | 400 | INVALID_ARGUMENT | Subscription Acknowledge (POST)
Invalid pull parameters | 400 | INVALID_ARGUMENT | Subscription Pull (POST)
Unauthorized | 401 | UNAUTHORIZED | All requests _(if a user is not authenticated)_
Forbidden Access to Resource  | 403 | FORBIDDEN | All requests _(if a user is forbidden to access the resource)_
