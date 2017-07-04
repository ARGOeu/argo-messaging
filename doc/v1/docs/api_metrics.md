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
         "resource_name": "host2.foo",
         "timeseries": [
            {
               "timestamp": "2017-07-04T09:36:03Z",
               "value": 50
            }
         ],
         "description": "Percentage value that displays the CPU usage of ams service in the specific node"
      },
   ]
}

```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
