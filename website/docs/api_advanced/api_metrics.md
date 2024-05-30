---
id: api_metrics
title: API Operational Metrics
sidebar_position: 6
---


Operational Metrics include metrics related to the CPU or memory usage of the ams nodes

## [GET] Get Operational Metrics

This request gets a list of operational metrics for the specific ams service

### Request

```
GET "/v1/metrics"
```

### Example request

```bash
curl -H "Content-Type: application/json" -H "x-api-key: S3CR3T"
 "https://{URL}/v1/metrics"
```

### Responses

If successful, the response returns a list of related operational metrics

Success Response
`200 OK`

```json
{
  "metrics": [
    {
      "metric": "ams_node.cpu_usage",
      "metric_type": "percentage",
      "value_type": "float64",
      "resource_type": "ams_node",
      "resource_name": "host.foo",
      "timeseries": [
        {
          "timestamp": "2017-07-04T10:18:07Z",
          "value": 0.2
        }
      ],
      "description": "Percentage value that displays the CPU usage of ams service in the specific node"
    },
    {
      "metric": "ams_node.memory_usage",
      "metric_type": "percentage",
      "value_type": "float64",
      "resource_type": "ams_node",
      "resource_name": "host.foo",
      "timeseries": [
        {
          "timestamp": "2017-07-04T10:18:07Z",
          "value": 0.1
        }
      ],
      "description": "Percentage value that displays the Memory usage of ams service in the specific node"
    }
  ]
}

```

### Errors

Please refer to section [Errors](/api_basic/api_errors.md) to see all possible Errors

## [GET] Get Health status

### Request

```
GET "/v1/status"
```

### Example request

```bash
curl -H "Content-Type: application/json" -H "x-api-key: S3CR3T"
 "https://{URL}/v1/status"
```

### Responses

If successful, the response returns the health status of the service

Success Response
`200 OK`

```json
{
  "status": "ok",
  "push_servers": [
    {
      "endpoint": "localhost:5555",
      "status": "Success: SERVING"
    }
  ]
}
```

### Errors

Please refer to section [Errors](/api_basic/api_errors.md) to see all possible Errors

## [GET] Get a VA Report

This request returns the total amount of messages per project for the given time window. The number of messages
is calculated using the `daily message count` for each one of the project's topics.

### Request

```
GET "/v1/metrics/daily-message-average"

```

### URL parameters

`start_date`: start date for querying projects topics daily message count(optional), default value is the start unix
time
`end_date`: start date for querying projects topics daily message count(optional), default is the time of the api call
`projects`: which projects to include to the query(optional), default is all registered projects

### Example request

```bash
curl -H "Content-Type: application/json" -H "x-api-key: S3CR3T"
 "https://{URL}/v1/metrics/va_metrics"
```

### Example request with URL parameters

```bash
curl -H "Content-Type: application/json" -H "x-api-key: S3CR3T"
 "https://{URL}/v1/metrics/va_metrics?start_date=2019-03-01&end_date=2019-07-24&projects=ARGO,ARGO-2"
```

### Responses

If successful, the response returns the total amount of messages per project for the given time window

Success Response
`200 OK`

```json
{
  "projects": [
    {
      "project": "ARGO-2",
      "message_count": 8,
      "average_daily_messages": 2
    },
    {
      "project": "ARGO",
      "message_count": 25669,
      "average_daily_messages": 120
    }
  ],
  "total_message_count": 25677,
  "average_daily_messages": 122
}
```

### Errors

Please refer to section [Errors](/api_basic/api_errors.md) to see all possible Errors

## [GET] Get User usage report

This is a combination of the va_metrics and the operational_metrics
api calls.The user will receive data for all of the projects that has
the project_admin role alongisde the operational metrics of the service.

### Request

```
GET "/v1/users/usageReport"

```

### URL parameters

`start_date`: start date for querying projects topics daily message count(optional), default value is the start unix
time
`end_date`: start date for querying projects topics daily message count(optional), default is the time of the api call
`projects`: which projects to include to the query(optional), default is all registered projects

### Example request

```bash
curl -H "Content-Type: application/json" -H "x-api-key: S3CR3T"
 "https://{URL}/v1/users/usageReport"
```

### Example request with URL parameters

```bash
curl -H "Content-Type: application/json" -H "x-api-key: S3CR3T"
 "https://{URL}/v1/users/usageReport?start_date=2019-03-01&end_date=2019-07-24&projects=ARGO,ARGO-2"
```

### Responses

Success Response
`200 OK`

```json
{
  "va_metrics": {
    "projects_metrics": {
      "projects": [
        {
          "project": "e2epush",
          "message_count": 27,
          "average_daily_messages": 0.03,
          "topics_count": 3,
          "subscriptions_count": 6,
          "users_count": 0
        }
      ],
      "total_message_count": 27,
      "average_daily_messages": 0.03
    },
    "total_users_count": 0,
    "total_topics_count": 3,
    "total_subscriptions_count": 6
  },
  "operational_metrics": {
    "metrics": [
      {
        "metric": "ams_node.cpu_usage",
        "metric_type": "percentage",
        "value_type": "float64",
        "resource_type": "ams_node",
        "resource_name": "test-MBP",
        "timeseries": [
          {
            "timestamp": "2022-09-13T09:39:56Z",
            "value": 0
          }
        ],
        "description": "Percentage value that displays the CPU usage of ams service in the specific node"
      },
      {
        "metric": "ams_node.memory_usage",
        "metric_type": "percentage",
        "value_type": "float64",
        "resource_type": "ams_node",
        "resource_name": "test-MBP",
        "timeseries": [
          {
            "timestamp": "2022-09-13T09:39:56Z",
            "value": 0.1
          }
        ],
        "description": "Percentage value that displays the Memory usage of ams service in the specific node"
      },
      {
        "metric": "ams_node.cpu_usage",
        "metric_type": "percentage",
        "value_type": "float64",
        "resource_type": "ams_node",
        "resource_name": "4",
        "timeseries": [
          {
            "timestamp": "2022-09-13T09:39:56Z",
            "value": 0
          }
        ],
        "description": "Percentage value that displays the CPU usage of ams service in the specific node"
      },
      {
        "metric": "ams_node.memory_usage",
        "metric_type": "percentage",
        "value_type": "float64",
        "resource_type": "ams_node",
        "resource_name": "4",
        "timeseries": [
          {
            "timestamp": "2022-09-13T09:39:56Z",
            "value": 0.1
          }
        ],
        "description": "Percentage value that displays the Memory usage of ams service in the specific node"
      }
    ]
  }
}
```

### Errors

Please refer to section [Errors](/api_basic/api_errors.md) to see all possible Errors


