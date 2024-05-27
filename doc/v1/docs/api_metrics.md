#Operational Metrics API Calls

Operational Metrics include metrics related to the CPU or memory usage of the ams nodes

## [GET] Get Operational Metrics
This request gets a list of operational metrics for the specific ams servcice

### Request
```
GET "/v1/metrics"
```


### Example request

```bash
curl -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
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
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Get VA Metrics

This request returns the total amount of messages per project for the given time window.The number of messages
is calculated using the `daily message count` for each one of the project's topics.
It also returns the amount of created `users`, `topics` and `subscriptions`
within the given time window.

### Request
```
GET "/v1/metrics/va_metrics"

```
### URL parameters
`start_date`: start date for querying projects topics daily message count(optional), default value is the start unix time
`end_date`: start date for querying projects topics daily message count(optional), default is the time of the api call
`projects`: which projects to include to the query(optional), default is all registered projects

### Example request

```bash
curl -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
 "https://{URL}/v1/metrics/va_metrics"
```

### Example request with URL parameters

```bash
curl -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
 "https://{URL}/v1/metrics/va_metrics?start_date=2019-03-01&end_date=2019-07-24&projects=ARGO,ARGO-2"
```

### Responses
If successful, the response returns the total amount of messages per project,
users,topics and subscriptions for the given time window

Success Response
`200 OK`

```json
{
  "projects_metrics": {
    "projects": [
        {
            "project": "ARGO-2",
            "message_count": 8,
            "average_daily_messages": 2,
            "users_count": 40,
            "topics_count": 30,
            "subscriptions_count": 100 
        },
        {
            "project": "ARGO",
            "message_count": 25669,
            "average_daily_messages": 120,
            "users_count": 4,
            "topics_count": 3,
            "subscriptions_count": 0 
        }
    ],
    "total_message_count": 25677,
    "average_daily_messages": 122
  },
  "total_users_count": 44,
  "total_topics_count": 33,
  "total_subscriptions_count": 100 
}
```
### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
