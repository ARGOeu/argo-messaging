# Subscriptions Api Calls

## [PUT] Manage Subscriptions - Create subscriptions  
This request creates a new subscription in a project with a PUT request

### Request
`PUT /v1/projects/{project_name}/subscriptions/{subscription_name}`

### Where
- Project_name: Name of the project to create
- subscription_name: The subscription name to create

### Example request
```json
curl -X PUT -H "Content-Type: application/json"  -d 'PUTBODY'
 "https://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine ?key=S3CR3T"`
```

### PUT  BODY
```json
{
 "topic": "projects/BRAND_NEW/topics/monitoring",
 "ackDeadlineSeconds":10
}
```

### Responses  

Success Response
`200 OK`
```json
{
 "name": "projects/BRAND_NEW/subscriptions/alert_engine",
 "topic": "projects/BRAND_NEW/topics/monitoring",
 "ackDeadlineSeconds": 10  
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Manage Subscriptions - List Subscriptions

This request lists all subscriptions  in a project with a GET  request
### Request
`GET /v1/projects/{project_name}/subscriptions`

### Where
- Project_name: Name of the project to list the subscriptions

### Example request
```json
curl -X PUT -H "Content-Type: application/json"
  "https://{URL}/v1/projects/BRAND_NEW/subscriptions?key=S3CR3T"
```


### Responses  
Success Response
`200 OK`

```json
 "subscriptions": [
 {
  "name": "projects/BRAND_NEW/subscriptions/alert_engine",
  "topic": "projects/BRAND_NEW/topics/monitoring",
  "pushConfig": {},
  "ackDeadlineSeconds": 10
 }
]
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Manage Subscriptions - Get a subscription's list of authorized users
This request returns a list of authorized users to consume from the subscription

### Request
```json
GET /v1/projects/{project_name}/subscriptions/{sub_name}:acl
```

### Where
- Project_name: Name of the project to get
- Sub_name: The subscription name

### Example request

```json
curl -H "Content-Type: application/json"  
 "https://{URL}/v1/projects/BRAND_NEW/subscriptions/subscription:acl?key=S3CR3T"`
```

### Responses  

Success Response
`200 OK`
```json
{
 "authorized_users": ["userC","userD"]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors


## [DELETE] Manage Subscriptions - Delete Subscriptions
This request deletes a subscription in a project with a DELETE request

### Request
`DELETE /v1/projects/{project_name}/subscriptions/{subscription_name}`

### Where
- Project_name: Name of the project
- subscription_name: The subscription name to delete

### Example request

```json
curl -X DELETE -H "Content-Type: application/json"  
http://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine?key=S3CR3T
```

### Responses  

Success Response
Code: `200 OK`, Empty response if successful.

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [POST] Modify Push Configuration
This request modifies the push configuration of a subscription

### Request
`POST /v1/projects/{project_name}/subscriptions/{subscription_name}:modifyPushConfig`

### Post body:
```
{
  "pushConfig": {  "pushEndpoint": "",
                   "retryPolicy": { "type": "linear", "period": 300 }
  }
}
```

### Where
- Project_name: Name of the project
- subscription_name: The subscription name to consume
- pushConfig: configuration including pushEndpoint for the remote endpoint to receive the messages. Also includes retryPolicy (type of retryPolicy and period parameters)


### Example request

```json
curl -X POST -H "Content-Type: application/json"  
-d POSTDATA http://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine:modifyPushConfig?key=S3CR3T"
```

### post body:
```
{
  "pushConfig": {"pushEndpoint": "host:example.com:8080/path/to/hook",
                 "retryPolicy":  { "type": "linear", "period": 300 }
  }
}
```

### Responses  

Success Response
Code: `200 OK`, Empty response if successful.

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [POST] Pull messages from a subscription (Consume)

This request consumes messages from a subscription in a project with a POST request

### Request
`POST /v1/projects/{project_name}/subscriptions/{subscription_name}:pull`

### Post body:
```json
{
 "maxMessages": "1"
}
```

### Where
- Project_name: Name of the project
- subscription_name: The subscription name to consume
- maxMessages: the max number of messages to consume
- returnImmediately: (true or false) to prevent the subscriber from waiting if the queue is currently empty. If not specified the default value is true.

 You can specify the max number of messages returned by one call by setting maxMessages field. By default, the server will keep the connection open until at least one message is received; you can optionally set the returnImmediately field to true to prevent the subscriber from waiting if the queue is currently empty.


### Example request

```json
curl -X POST -H "Content-Type: application/json"
  -d POSTDATA https://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine:pull?key=S3CR3T"
```

### post body:
{
 "maxMessages": "1"
}

### Responses  

`200 OK`
```json
{
  "receivedMessages": [
    {
      "ackId": "dQNNHlAbEGEIBE...",
      "message": {
        "attributes": [
          {
            "key": "whatever",
            "value": "foo"
          }
        ],
        "data": "U28geW91IHdlbnQgYWhlYWQgYW5kIGRlY29kZWQgdGhpcywgeW91IGNvdWxkbid0IHJlc2lzdCBlaCA/",
        "messageId": "100309303"
      }
    }
  ]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors


## [POST] Sending an ACK
Messages retrieved from a pull subscription can be acknowledged by sending message with an array of ackIDs.

### Request
`POST /v1/projects/{project_name}/subscriptions/{subscription_name}:acknowledge`

### Post body:
```json
{
  "ackIds": [
  "dQNNHlAbEGEIBE..."
 ]

}
```

### Where
- Project_name: Name of the project
- subscription_name: The subscription name to consume
- ackIds: the ids of the messages


### Example request


```
curl -X POST -H "Content-Type: application/json"  
-d POSTDATA http://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine:acknowledge?key=S3CR3T"
```

### post body:
```json
{
 "ackIds": [
  "dQNNHlAbEGEIBE..."
 ]
}
```

### Responses  
Success Response
`200 OK`

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Get Offsets
This request returns the min, max and current offset of a subscription

### Request
`GET /v1/projects/{project_name}/subscriptions/{subscription_name}:Offsets`

### Post body:
```
{
  "max": 14,
  "min": 0,
  "current": 4
}
```

### Where
- Project_name: Name of the project
- subscription_name: The subscription name to consume


### Example request

```json
curl -X GET -H "Content-Type: application/json"  
-d POSTDATA http://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine:offsets?key=S3CR3T"
```

### post body:
```
{
  "max": 14,
  "min": 0,
  "current": 4
}
```

### Responses  

Success Response
Code: `200 OK`, Empty response if successful.

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [POST] Modify Offsets
This request modifies the current offset of a subscription

### Request
`POST /v1/projects/{project_name}/subscriptions/{subscription_name}:modifyOFfset`

### Post body:
```
{
 "offset":3
}
```

### Where
- Project_name: Name of the project
- subscription_name: The subscription name to consume
- offset_config: an offset number in int64



### Example request

```json
curl -X POST -H "Content-Type: application/json"  
-d POSTDATA http://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine:modifyOffset?key=S3CR3T"
```

### post body:
```
{
  "offset":14
}
```

### Responses  

Success Response
Code: `200 OK`, Empty response if successful.

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Subscription Metrics
The following request returns related metrics for the specific subscription: for eg the number of consumed messages

### Request
```
GET "/v1/projects/{project_name}/subscriptions/{sub_name}:metrics"
```

### Where
- Project_name: name of the project
- sub_name: name of the subscription

### Example request

```json
curl  -H "Content-Type: application/json"
"https://{URL}/v1/projects/BRAND_NEW/subscriptions/monitoring:metrics?key=S3CR3T"
```

### Responses  
If successful it returns the number of messages consumed in the specific subscription
Success Response
`200 OK`
```
{
   "metrics": [
      {
         "metric": "subscription.number_of_messages",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "subscription",
         "resource_name": "sub1",
         "timeseries": [
            {
               "timestamp": "2017-06-30T14:20:38Z",
               "value": 0
            }
         ],
         "description": "Counter that displays the number number of messages published to the specific topic"
      },
      {
         "metric": "topic.number_of_bytes",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "topic",
         "resource_name": "sub1",
         "timeseries": [
            {
               "timestamp": "2017-06-30T14:20:38Z",
               "value": 0
            }
         ],
         "description": "Counter that displays the total size of data (in bytes) published to the specific topic"
      }
   ]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
