#Operational Metrics API Calls

Operational Metrics include metrics related to the CPU or memory usage of the ams nodes

## [GET] Get Operational Metrics
This request gets a list of operational metrics for the specific ams servcice

### Request
```json
GET "/v1/metrics"
```


### Example request

```json
curl -H "Content-Type: application/json"
 "https://{URL}/v1/metrics?key=S3CR3T"
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

## [GET] Get Health status

### Request
```
GET "/v1/status"
```

### Example request

```
curl -H "Content-Type: application/json"
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