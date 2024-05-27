# Subscriptions Api Calls

## [PUT] Manage Subscriptions - Create subscriptions  
This request creates a new subscription in a project with a PUT request

### Request
`PUT /v1/projects/{project_name}/subscriptions/{subscription_name}`

### Where
- Project_name: Name of the project to create
- subscription_name: The subscription name to create

### Example request
```bash
curl -X PUT -H "Content-Type: application/json" -H "x-api-token:S3CR3T"  -d 'PUTBODY'
 "https://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine"
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
 "ackDeadlineSeconds": 10  ,
 "createdOn": "2020-11-19T00:00:00Z"
}
```

### Push Enabled Subscriptions
Whenever a subscription is created with a valid push configuration, the service will also generate a unique hash that
should be later used to validate the ownership of the registered push endpoint, and will mark the subscription as 
unverified.

The `type` field specifies what kind of push subscription the service will handle.
The `http_endpoint` type is about subscriptions that will forward their messages
to remote http endpoints. The `mattermost` type is about subscriptions that will
forward their messages to mattermost channels through a mattermost webhook.

The `maxMessages` field declares the number of messages that should be sent per
push action. The default value is `1`. If `maxMessages` holds a value of `1` your
push endpoint should expect a request body with the following schema:

The `base64Decode` field indicates that the push mechanism should
decode each message before sending it to the remote destination.

```json
{
     "message": {
       "attributes": {
         "key": "value"
       },
       "data": "SGVsbG8gQ2xvdWQgUHViL1N1YiEgSGVyZSBpcyBteSBtZXNzYWdlIQ==",
       "messageId": "136969346945"
     },
     "subscription": "projects/myproject/subscriptions/mysubscription"
   }
```

If the `maxMessages` field holds a value of greater than `1` your push endpoint
should expect a request body with the following schema:
```json
{  
   "messages":[  
      {  
         "message":{  
            "attributes":{  
               "key":"value"
            },
            "data":"SGVsbG8gQ2xvdWQgUHViL1N1YiEgSGVyZSBpcyBteSBtZXNzYWdlIQ==",
            "messageId":"136969346945"
         },
         "subscription":"projects/myproject/subscriptions/mysubscription"
      },
      {  
         "message":{  
            "attributes":{  
               "key":"value"
            },
            "data":"SGVsbG8gQ2xvdWQgUHViL1N1YiEgSGVyZSBpcyBteSBtZXNzYWdlIQ==",
            "messageId":"136969346945"
         },
         "subscription":"projects/myproject/subscriptions/mysubscription"
      }
   ]
}
```

## Request to create Push Enabled Subscription for http_endpoint
```json
{
 "topic": "projects/BRAND_NEW/topics/monitoring",
 "ackDeadlineSeconds":10,
  "pushConfig": {
    "type": "http_endpoint",
    "pushEndpoint": "https://127.0.0.1:5000/receive_here",
    "maxMessages": 3,
    "retryPolicy": {
      "type": "linear", 
      "period": 1000              	
    }
   }
}
```
### Response
```json
{
 "name": "projects/BRAND_NEW/subscriptions/alert_engine",
 "topic": "projects/BRAND_NEW/topics/monitoring",
 "ackDeadlineSeconds": 10,
  "pushConfig": {
    "pushEndpoint": "https://127.0.0.1:5000/receive_here",
    "maxMessages": 3,
    "authorizationHeader": {
      "type": "autogen",
      "value": "4551h9j7f7dde380a5f8bc4fdb4fe980c565b67b"
    } ,
    "retryPolicy": {
      "type": "linear", 
      "period": 1000              	
    },
    "verificationHash": "9d5189f7f758e380a5f8bc4fdb4fe980c565b67b",
    "verified": false,
    "mattermostUrl": "",
    "mattermostUsername": "",
    "mattermostChannel": ""
    },
  "createdOn": "2020-11-19T00:00:00Z"
}
```


## Request to create Push Enabled Subscription for mattermost
```json
{
 "topic": "projects/BRAND_NEW/topics/monitoring",
 "ackDeadlineSeconds":10,
  "pushConfig": {
    "type": "mattermost",
    "mattermostUrl": "webhook.com",
    "mattermostUsername": "mattermost",
    "mattermostChannel": "channel",
    "retryPolicy": {
      "type": "linear", 
      "period": 1000              	
    }
   }
}
```

### Response
```json
{
 "name": "projects/BRAND_NEW/subscriptions/alert_engine",
 "topic": "projects/BRAND_NEW/topics/monitoring",
 "ackDeadlineSeconds": 10,
  "pushConfig": {
    "pushEndpoint": "",
    "maxMessages": 1,
    "retryPolicy": {
      "type": "linear", 
      "period": 1000              	
    },
    "verificationHash": "",
    "verified": true,
    "mattermostUrl": "webhook.com",
    "mattermostUsername": "mattermost",
    "mattermostChannel": "channel"
    },
  "createdOn": "2020-11-19T00:00:00Z"
}
```

### Authorization headers

Specify an `authorization header` value and how it is going to be generated,
to be included in the outgoing push request with each message, to the remote
push endpoint.
 
- `autogen(default)`: The authorization header value will be automatically
generated by the service itself.
- `disabled`: No authorization header will be provided with the outgoing
push requests.

### Different Retry Policies
Creating a push enabled subscription with a `linear` retry policy and a `period` of 3000 means that you will be receiving
message(s) every `3000ms`.

If you decide to choose a retry policy of `slowstart`, you will be receiving messages with dynamic internals.
The `slowstart` retry policy starts by pushing the first message(s) and then deciding the time that should elapse 
before the next push action.
- `IF` the message(s) are delivered successfully the elapsed time until the next push request will be halved, until it reaches
the lower limit of `300ms`.

- `IF` the message(s) are not delivered successfully the elapsed time until the next push request will be doubled, until 
it reached the upper limit of `1day`.

So for example, the first push action will have by default a `1 second` interval. If it successful the next push re request will
happen in `0.5 seconds`. If it is unsuccessful the next push request will happen in `2 seconds`.


### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors


## [POST] Manage Subscriptions - Verify ownership of a push endpoint
This request triggers the process of verifying the ownership of a registered push endpoint 

### Request
`PUT /v1/projects/{project_name}/subscriptions/{subscription_name}:verifyPushEndpoint`

### Where
- Project_name: Name of the project
- subscription_name: The subscription name

### Example request
```bash
curl -X POST  -H "x-api-token:S3CR3T" 
"https://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine:verifyPushEndpoint"
```

### Push Enabled Subscriptions
Whenever a subscription is created with a valid push configuration, the service will also generate a unique hash that
should be later used to validate the ownership of the registered push endpoint, and will mark the subscription as 
unverified.

The owner of the push endpoint needs to execute the following steps in order to verify the ownership of the
registered endpoint.

- Open an api call with a path of `/ams_verificationHash`. The service will try to access this path using the `host:port`
of the push endpoint. For example, if the push endpoint is `https://example.com:8443/receive_here`, the  push endpoint should also
support the api route of `https://example.com:8443/ams_verificationHash`.

- The api route of `https://example.com:8443/ams_verificationHash` should support the http `GET` method.

- A `GET` request to `https://example.com:8443/ams_verificationHash` should return a response body 
with only the `verificationHash`
that is found inside the subscriptions push configuration, 
a `status code` of `200` and the header `Content-type: plain/text`.

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Manage Subscriptions - List All Subscriptions under a specific Topic

This request lists all available subscriptions under a specific topic in the service.

### Request
`GET /v1/projects/{project_name}/topics/{topic_name}/subscriptions`

### Where
- Project_name: Name of the project the topic belongs to
- Topic_name: Name of the topic

### Example request
```bash
curl -X GET -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
  "https://{URL}/v1/projects/p1/topics/t1/subscriptions"
```

Success Response
`200 OK`

```json
{
 "subscriptions": [
 "/projects/p1/subscriptions/sub1",
 "/projects/p1/subscriptions/sub2"
 ]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Manage Subscriptions - List All Subscriptions

This request lists all available subscriptions under a specific project in the service using pagination

If the `USER` making the request has only `consumer` role for the respective project, it will load
only the subscriptions that he has access to(being present in a subscriptions's acl).

It is important to note that if there are no results to return the service will return the following:

Success Response
`200 OK`

```json
{
 "subscriptions": [],
  "nextPageToken": "",
  "totalSize": 0
 }
```
Also the default value for `pageSize = 0` and `pageToken = "`.

`Pagesize = 0` returns all the results.

### Paginated Request that returns all subscriptions under the specified project

This request lists all subscriptions  in a project with a GET  request
### Request
`GET /v1/projects/{project_name}/subscriptions`

### Where
- Project_name: Name of the project to list the subscriptions

### Example request
```bash
curl -X GET -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
  "https://{URL}/v1/projects/BRAND_NEW/subscriptions"
```


### Responses  
Success Response
`200 OK`

```json
 {
  "subscriptions":[
  {
    "name": "projects/BRAND_NEW/subscriptions/alert_engine",
    "topic": "projects/BRAND_NEW/topics/monitoring",
    "pushConfig": {},
    "ackDeadlineSeconds": 10,
    "createdOn": "2020-11-19T00:00:00Z"
  },
 {
   "name": "projects/BRAND_NEW/subscriptions/alert_engine2",
   "topic": "projects/BRAND_NEW/topics/monitoring",
   "pushConfig": {},
   "ackDeadlineSeconds": 10,
   "createdOn": "2020-11-19T00:00:00Z"
 }],
 "nextPageToken": "",
 "totalSize": 2
}
```

### Paginated Request that returns the next page of a specific size

This request lists subscriptions  in a project with a GET  request
### Request
`GET /v1/projects/{project_name}/subscriptions`

### Where
- Project_name: Name of the project to list the subscriptions

### Example request
```bash
curl -X PUT -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
  "https://{URL}/v1/projects/BRAND_NEW/subscriptions?pageSize=1&pageToken=some_token"
```


### Responses  
Success Response
`200 OK`

```json
 {
  "subscriptions":[
   {
    "name": "projects/BRAND_NEW/subscriptions/alert_engine",
    "topic": "projects/BRAND_NEW/topics/monitoring",
    "pushConfig": {},
    "ackDeadlineSeconds": 10,
    "createdOn": "2020-11-19T00:00:00Z"
  }
 ],
 "nextPageToken": "",
 "totalSize": 2
}
```

### Paginated Request that returns the first page of a specific size

This request lists subscriptions  in a project with a GET  request
### Request
`GET /v1/projects/{project_name}/subscriptions`

### Where
- Project_name: Name of the project to list the subscriptions

### Example request
```bash
curl -X PUT -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
  "https://{URL}/v1/projects/BRAND_NEW/subscriptions?pageSize=1"
```


### Responses  
Success Response
`200 OK`

```json
 {
 "subscriptions":[
  {
    "name": "projects/BRAND_NEW/subscriptions/alert_engine2",
    "topic": "projects/BRAND_NEW/topics/monitoring",
    "pushConfig": {},
    "ackDeadlineSeconds": 10,
    "createdOn": "2020-11-19T00:00:00Z"
  }
 ],
 "nextPageToken": "some_token",
 "totalSize": 2
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Manage Subscriptions - Get a subscription's list of authorized users
This request returns a list of authorized users to consume from the subscription

### Request
```
GET /v1/projects/{project_name}/subscriptions/{sub_name}:acl
```

### Where
- Project_name: Name of the project to get
- Sub_name: The subscription name

### Example request

```bash
curl -H "Content-Type: application/json"   -H "x-api-token:S3CR3T" 
 "https://{URL}/v1/projects/BRAND_NEW/subscriptions/subscription:acl"
```

### Responses  

Success Response
`200 OK`
```
{
 "authorized_users": ["userC","userD"]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [POST] Modify ACL of a given subscription
The following request Modifies the authorized users list of a given subscription

### Request
```
POST "/v1/projects/{project_name}/subscriptions/{sub_name}:modifyAcl"
```

### Where
- project_name: Name of the project
- sub_name: name of the subscription


### Post data
```
{
"authorized_users": [
 "UserX","UserY"
]
}
```

### Example request

```bash
curl -X POST -H "Content-Type: application/json"    -H "x-api-token:S3CR3T" 
-d $POSTDATA "https://{URL}/v1/projects/BRAND_NEW/subscriptions/subscription:modifyAcl"
```

### Responses  

Success Response
`200 OK`

### Errors
If the to-be updated ACL contains users that are non-existent in the project, the API returns the following error:
`404 NOT_FOUND`
```
{
   "error": {
      "code": 404,
      "message": "User(s): UserFoo1,UserFoo2 do not exist",
      "status": "NOT_FOUND"
   }
}
```

Please refer to section [Errors](api_errors.md) to see all possible Errors

## [DELETE] Manage Subscriptions - Delete Subscriptions
This request deletes a subscription in a project with a DELETE request

### Request
`DELETE /v1/projects/{project_name}/subscriptions/{subscription_name}`

### Where
- Project_name: Name of the project
- subscription_name: The subscription name to delete

### Example request

```bash
curl -X DELETE -H "Content-Type: application/json"   -H "x-api-token:S3CR3T" 
"http://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine"
```

### Responses  

Success Response
Code: `200 OK`, Empty response if successful.

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [POST] Modify Ack Deadline
This request modifies the acknowledgment deadline for the subscription. The ack deadline value is measured in seconds. The minimum ack deadline value allowed is 0sec and the maximum 600sec.

### Request
`POST /v1/projects/{project_name}/subscriptions/{subscription_name}:modifyAckDeadline`

### Post body:
```
{
  "ackDeadlineSeconds": 20
}
```

### Where
- Project_name: Name of the project
- subscription_name: The subscription name to consume
- ackDeadlineSeconds: integer representing seconds for the acknowledgment deadline (min=0sec, max=600sec).


### Example request

```bash
curl -X POST -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
-d $POSTDATA "http://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine:modifyAckDeadline"
```

### post body:
```
{
  "ackDeadlineSeconds": 30
}
```

### Responses  

Success Response
Code: `200 OK`, Empty response if successful. The deadline will change to 30seconds

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [POST] Modify Push Configuration
This request modifies the push configuration of a subscription

### Request
`POST /v1/projects/{project_name}/subscriptions/{subscription_name}:modifyPushConfig`

### Post body for http_endpoint
```json
{  
   "pushConfig":{  
      "type": "http_endpoint",
      "pushEndpoint":"example.com",
      "maxMessages": 5,
      "authorizationHeader": {
         "type": "autogen"
      },
      "retryPolicy":{  
         "type":"linear",
         "period":300
      },
      "base64Decode": false
   }
}
```

### Post body for mattermost
```json
{  
   "pushConfig":{  
      "type": "mattermost",
      "retryPolicy":{  
         "type":"linear",
         "period":300
      },
      "mattermostUrl": "webhook.com",
      "mattermostUsername": "willy",
      "mattermostChannel": "ops",
      "base64Decode": true
   }
}
```

### Where
- Project_name: Name of the project
- subscription_name: The subscription name to consume
- pushConfig: configuration including pushEndpoint for the remote endpoint to receive the messages. Also includes retryPolicy (type of retryPolicy and period parameters)

- `autogen(default when modyfing a sub from pull to push)`: The authorization header value will be automatically
generated by the service itself.
- `disabled`: No authorization header will be provided with the outgoing
push requests.

<b>NOTE</b> that if you updated a push configuration with <b>autogen</b>
the service will generate a new value every time the update request happens.
For example, if you want to update your authorization header value,
you can use the update request with the <b>autogen</b> type.

### Example request

```bash
curl -X POST -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
-d $POSTDATA "http://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine:modifyPushConfig"
```

### post body:
```json
{  
   "pushConfig":{  
      "type": "http_endpoint",
      "pushEndpoint":"host:example.com:8080/path/to/hook",
      "maxMessages": 3,
      "retryPolicy":{  
         "type":"linear",
         "period":300
      }
   }
}
```

### Responses  

Success Response
Code: `200 OK`, Empty response if successful.

Whenever a subscription is created with a valid push configuration, the service will also generate a unique hash that
should be later used to validate the ownership of the registered push endpoint, and will mark the subscription as 
unverified.

**NOTE** Changing the push endpoint of a push enabled subscription, or removing the push configuration and then re-applying
will mark the subscription as unverified and a new verification process should take place.

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [POST] Pull messages from a subscription (Consume)

This request consumes messages from a subscription in a project with a POST request.

It's important to note that the subscription's topic must exist in order for the user to pull messages.

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

```bash
curl -X POST -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
  -d $POSTDATA https://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine:pull"
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


```bash
curl -X POST -H "Content-Type: application/json"   -H "x-api-token:S3CR3T" 
-d $POSTDATA http://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine:acknowledge"
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
`GET /v1/projects/{project_name}/subscriptions/{subscription_name}:offsets`

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

```bash
curl -X GET -H "Content-Type: application/json"   -H "x-api-token:S3CR3T" 
-d $POSTDATA http://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine:offsets"
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

## [GET] Get Offset by Timestamp
This request returns the offset of the first message with a timestamp equal or greater than the time given.

### Request
`GET /v1/projects/{project_name}/subscriptions/{subscription_name}:timeToOffset?time={{timestamp}}`

### Where
- Project_name: Name of the project
- subscription_name: The subscription name to consume
- timestamp: timestamp in `Zulu` format - `(2006-11-02T13:39:11.000Z)`

### Example request

```bash
curl -X GET -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
"http://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine:timeToOffset?time=2019-09-02T13:39:11.100Z"
```

### Responses  

Success Response
Code: `200 OK`

### Response body:
```
{
  "offset": 640
}
```

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

```bash
curl -X POST -H "Content-Type: application/json" -H "x-api-token:S3CR3T" 
-d $POSTDATA "http://{URL}/v1/projects/BRAND_NEW/subscriptions/alert_engine:modifyOffset"
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

```bash
curl  -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
"https://{URL}/v1/projects/BRAND_NEW/subscriptions/monitoring:metrics"
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
      },
      {
         "metric": "subscription.consumption_rate",
         "metric_type": "rate",
         "value_type": "float64",
         "resource_type": "subscription",
         "resource_name": "sub1",
         "timeseries": [
            {
               "timestamp": "2019-05-06T00:00:00Z",
               "value": 10
            }
         ],
         "description": "A rate that displays how many messages were consumed per second between the last two consume events"
      }
   ]
}
```



### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
