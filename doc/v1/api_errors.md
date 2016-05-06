---
title: 'ARGO Messaging Service documentation | ARGO'
page_title: ARGO Messaging API Error Messages
font_title: fa fa-cogs
description: ARGO Messaging API Error Messages
---

# Errors

In case of Error during handling user’s request the API responds using the following schema
```json
{
   "error": {
      "code": 500,
      "message": "Something bad happened",
      "errors": [
         {
            "message": "Something bad happened",
            "domain": "global",
            "reason": "backend"
         }
      ],
      "status": "INTERNAL"
   }
}
```
Most of the times the errors array is empty thus omitted such as:
```json
{
   "error": {
      "code": 500,
      "message": "Something bad happened",
      "status": "INTERNAL"
   }
}
```
Captured Errors from usage scenarios
Scenario | Error Message
-------- | -------------
Put topic with the same name |
```json
{
  "error": {
    "code": 409,
    "message": "Topic already exists",
    "status": "ALREADY_EXISTS"
  }
}
```
Put subscription with the same name |
```json
{
  "error": {
    "code": 409,
    "message": "Subscription already exists",
    "status": "ALREADY_EXISTS"
  }
}
```
Invalid Topics name |
```json
{
  "error": {
    "code": 400,
    "message": "Invalid topics name",
    "status": "INVALID_ARGUMENT"
  }
}
```
Get a subscription that doesn’t exist |
```json
{
  "error": {
    "code": 404,
    "message": "Subscription does not exist",
    "status": "NOT_FOUND"
  }
}
```
