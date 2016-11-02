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
	" https://{URL}/v1/projects/EGI/subscriptions/alert_engine ?key=S3CR3T"`
```

### PUT  BODY
```json
{
 "topic": "projects/EGI/topics/monitoring",
 "ack":10
}
```

### Responses  

Success Response
`200 OK`
```json
{
 "name": "projects/EGI/subscriptions/alert_engine",
 "topic": "projects/EGI/topics/monitoring",
 "ackDeadlineSeconds": 10  
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Manage Subscriptions - List Subscriptions

This request lists all subscriptions  in a project with a GET  request
### Request
`GET /v1/projects/{project_name}/subscriptions/`

### Where
- Project_name: Name of the project to list the subscriptions

### Example request
```json
curl -X PUT -H "Content-Type: application/json"
  -d '' " https://{URL}/v1/projects/EGI/subscriptions/?key=S3CR3T"
```


### Responses  
Success Response
`200 OK`

```json
 "subscriptions": [
 {
  "name": "projects/EGI/subscriptions/alert_engine",
  "topic": "projects/EGI/topics/monitoring",
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
`GET /v1/projects/{project_name}/subscriptions/{sub_name}:acl``

### Where
- Project_name: Name of the project to get
- Sub_name: The subscription name

### Example request

```json
curl -H "Content-Type: application/json"  
-d '' " https://{URL}/v1/projects/EGI/subscriptions/subscription:acl?key=S3CR3T"`
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
This request deletes a topic in a project with a DELETE request

### Request
`DELETE /v1/projects/{project_name}/subscriptions/{subscription_name}`

### Where
- Project_name: Name of the project
- subscription_name: The subscription name to delete

### Example request

```json
curl -X DELETE -H "Content-Type: application/json"  
http://{URL}/v1/projects/EGI/subscriptions/alert_engine?key=S3CR3T
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
json
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

`curl -X POST -H "Content-Type: application/json"  
-d POSTDATA http://{URL}/v1/projects/EGI/subscriptions/alert_engine:modifyPushConfig?key=S3CR3T"
`

### post body:
```json
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


### Example request

```json
curl -X POST -H "Content-Type: application/json"
  -d POSTDATA https://{URL}/v1/projects/EGI/subscriptions/alert_engine:pull?key=S3CR3T"
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
Messages retrieved from a pull subscription can be acknowledged by sending message with an array of ackIDs. In the current implementation, the service will retrieve the ackID corresponding to the highest message offset and will consider that message and all previous messages as acknowledged by the consumer.

### Request
`POST /v1/projects/{project_name}/subscriptions/{subscription_name}:acknowledge`

### Post body:
```json
{
  "ackIds": [
  "dQNNHlAbEGEIBE..."
 ],

}
```

### Where
- Project_name: Name of the project
- subscription_name: The subscription name to consume
- ackIds: the ids of the messages


### Example request


```
curl -X POST -H "Content-Type: application/json"  
-d POSTDATA http://{URL}/v1/projects/EGI/subscriptions/alert_engine:acknowledge?key=S3CR3T"
```

### post body:
```json
{
 "ackIds": [
  "dQNNHlAbEGEIBE..."
 ],

}
```

### Responses  
Success Response
`200 OK`

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
