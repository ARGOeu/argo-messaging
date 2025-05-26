---
id: api_version
title: Get API Version information
sidebar_position: 8
---


This method can be used to retrieve api version information

## Input

```
GET /v1/version
```

### Request headers

```
Accept: application/json
```

## Response

Headers: `Status: 200 OK`

## Response Body

Json Response

```json
{
  "release": "1.0.5",
  "commit": "f9f2e8c5f02lbcc94fe76b0d3cfa5d20d9365444",
  "build_time": "2019-11-01T12:51:04Z",
  "golang": "go1.11.5",
  "compiler": "gc",
  "os": "linux",
  "architecture": "amd64"
}
```
