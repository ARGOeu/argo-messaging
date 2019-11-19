#Topics Api Calls

Topics are resources that can hold messages. Publishers (users/systems) can create topics on demand and name them (Usually with names that make sense and express the class of messages delivered in the topic).
A topic name must be scoped to a project.

## [PUT] Manage Topics - Create new topic
This request creates a new topic with the given topic_name in a project with a PUT request

### Request
```json
PUT "/v1/projects/{project_name}/topics/{topic_name}"
```

### RequestBody
If you need to link a schema with the topic you need to provide its name. 
```json
{
  "schema": "schema-1"
}
```

### Where
- Project_name: Name of the project to create
- Topic_name: The topic name to create

### Example request
```json
curl -X PUT -H "Content-Type: application/json"
 " https://{URL}/v1/projects/BRAND_NEW/topics/monitoring?key=S3CR3T"
```

### Responses  

If successful, the response contains the newly created topic.

Success Response
`200 OK`
```json
{
 "name": "projects/BRAND_NEW/topics/monitoring"
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors


## [DELETE] Manage Topics - Delete topic
This request deletes the defined topic in a project with a DELETE request

### Request
```json
DELETE "/v1/projects/{project_name}/topics/{topic_name}"
```

### Where
- Project_name: Name of the project to delete
- Topic_name: The topic name to delete

### Example request

```json
curl -X DELETE -H "Content-Type: application/json"  
-d '' "https://{URL}/v1/projects/BRAND_NEW/topics/monitoring?key=S3CR3T"
```

### Responses  

Success Response
Code: `200 OK`, Empty response if successful.

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Manage Topics - Get a topic
This request gets the details of a topic in a project with a GET request

### Request
```json
GET "/v1/projects/{project_name}/topics/{topic_name}"
```

### Where
- Project_name: Name of the project to get
- Topic_name: The topic name to get

### Example request

```json
curl -H "Content-Type: application/json"  
 "https://{URL}/v1/projects/BRAND_NEW/topics/monitoring?key=S3CR3T"
```

### Responses  
If successful, the response returns the details of the defined topic.

Success Response
`200 OK`
```json
{
 "name": "projects/BRAND_NEW/topics/monitoring"
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Manage Topics - List Topics
This request lists all available topics under a specific project in the service using pagination.

If the `USER` making the request has only `publisher` role for the respective project, it will load
only the topics that he has access to(being present in a topic's acl).

It is important to note that if there are no results to return the service will return the following:

Success Response
`200 OK`

```json
{
 "users": [],
  "nextPageToken": "",
  "totalSize": 0
 }
```
Also the default value for `pageSize = 0` and `pageToken = "`.

`Pagesize = 0` returns all the results.

### Paginated Request that returns all topics under the specified project

```GET "/v1/projects/{project_name}/topics"```

### Where
 - Project_name: Name of the project to get the list of topics

### Example request

```
curl -H "Content-Type: application/json"
"https://{URL}/v1/projects/BRAND_NEW/topics/?key=S3CR3T"`
```

### Responses  

Success Response
`200 OK`
```json
{
  "topics": [
    {
      "name":"/project/BRAND_NEW/topics/monitoring"
    },
    {
      "name":"/project/BRAND_NEW/topics/accounting"
    }
 ],
  "nextPageToken": "",
  "totalSize": 2
}
```

### Paginated Request that returns the first page of a specific size

```GET "/v1/projects/{project_name}/topics"```


### Where
 - Project_name: Name of the project to get the list of topics

### Example request

```
curl -H "Content-Type: application/json"
"https://{URL}/v1/projects/BRAND_NEW/topics/?key=S3CR3T&pageSize=1"`
```

### Responses  

Success Response
`200 OK`
```json
{
  "topics": [
    {
      "name":"/project/BRAND_NEW/topics/monitoring"
    }
 ],
  "nextPageToken": "some_token",
  "totalSize": 2
}
```

### Paginated Request that returns the next  page of a specific size

```GET "/v1/projects/{project_name}/topics"```


### Where
 - Project_name: Name of the project to get the list of topics

### Example request

```
curl -H "Content-Type: application/json"
"https://{URL}/v1/projects/BRAND_NEW/topics/?key=S3CR3T&pageSize=1&pageToken=some_token"`
```

### Responses  

Success Response
`200 OK`
```json
{
  "topics": [
    {
      "name":"/project/BRAND_NEW/topics/accounting"
    }
 ],
  "nextPageToken": "",
  "totalSize": 2
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [POST] Publish message/s to a topic
The topic:publish endpoint publishes a message, or a list of messages to a specific topic with a  POST request

### Request
```json
POST "/v1/projects/{project_name}/topics/{topic_name}:publish"
```

### Where
- Project_name: Name of the project to post the messages
- topic_name: to post the messages


### Post data
```json
{
"messages": [
 	{
  		"attributes": {
        "attr1":"test1",
        "attr2":"test2"
   		}
  	,
 "data":"U28geW91IHdlbnQgYWhlYWQgYW5kIGRlY29kZWQgdGhpcywgeW91IGNvdWxkbid0IHJlc2lzdCBlaCA/"

 	}
]
}
```

> The value of the data property must be always encoded in base64 format.


### Example request

```json
curl -X POST -H "Content-Type: application/json"  
-d { POSTDATA } "https://{URL}/v1/projects/BRAND_NEW/topics/monitoring:publish?key=S3CR3T"
```

### Responses  

If successful, the response contains the messageIds of the messages published.

Success Response `200 OK`
```json
{
 "messageIds": [
  "100309303"
 ]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors


## [GET] List ACL of a given topic
The following request returns a list of authorized users (publishers) of a given topic.

### Request
```
GET "/v1/projects/{project_name}/topics/{topic_name}:acl"
```

### Where
- Project_name: name of the project
- topic_name: name of the topic

### Example request

```json
curl  -H "Content-Type: application/json"
"https://{URL}/v1/projects/BRAND_NEW/topics/monitoring:acl?key=S3CR3T"
```

### Responses  
If successful it returns the authorized users of the topic.

Success Response
`200 OK`
```
{
 "authorized_users": [
  "UserA","UserB"
 ]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors



## [POST] Modify ACL of a given topic
The following request Modifies the authorized users list of a given topic

### Request
```
POST "/v1/projects/{project_name}/topics/{topic_name}:modifyAcl"
```

### Where
- Project_name: Name of the project
- topic_name: name of the topic


### Post data
```
{
"authorized_users": [
 "UserX","UserY"
]
}
```

### Example request

```
curl -X POST -H "Content-Type: application/json"  
-d { POSTDATA } "https://{URL}/v1/projects/BRAND_NEW/topics/monitoring:modifyAcl?key=S3CR3T"
```

### Responses  

Success Response
`200 OK`

### Errors
If the to-be updated ACL contains users that are non-existent in the project the API returns the following error:
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

## [GET] Topic Metrics
The following request returns related metrics for the specific topic: for eg the number of published messages

### Request
```
GET "/v1/projects/{project_name}/topics/{topic_name}:metrics"
```

### Where
- Project_name: name of the project
- topic_name: name of the topic

### Example request

```json
curl  -H "Content-Type: application/json"
"https://{URL}/v1/projects/BRAND_NEW/topics/monitoring:metrics?key=S3CR3T"
```

### Responses  
If successful it returns topic's related metrics (number of messages published and total bytes).

Success Response
`200 OK`
```
{
   "metrics": [
      {
         "metric": "topic.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "2017-06-27T10:20:18Z",
               "value": 1
            }
         ],
         "description": "Counter that displays the number of subscriptions belonging to a specific topic"
      },
      {
         "metric": "topic.number_of_messages",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "2017-06-27T10:20:18Z",
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
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "2017-06-27T10:20:18Z",
               "value": 0
            }
         ],
         "description": "Counter that displays the total size of data (in bytes) published to the specific topic"
      },
      {
         "metric": "topic.number_of_daily_messages",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "2018-10-02",
               "value": 30
            },
            {
               "timestamp": "2018-10-01",
               "value": 40
            }
         ],
         "description": "A collection of counters that represents the total number of messages published each day to a specific topic"
      },
      {
         "metric": "topic.publishing_rate",
         "metric_type": "rate",
         "value_type": "float64",
         "resource_type": "topic",
         "resource_name": "topic1",
         "timeseries": [
            {
               "timestamp": "2019-05-06T00:00:00Z",
               "value": 10
            }
         ],
         "description": "A rate that displays how many messages were published per second between the last two publish events"
      }
   ]
}

```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
