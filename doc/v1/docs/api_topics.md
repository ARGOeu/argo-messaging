---
title: 'ARGO Messaging Service documentation | ARGO'
page_title: Argo Messaging API Topics related Calls
font_title: fa fa-cogs
description: Argo Messaging API Topics related Calls
---

# Manage Topics - Create new topic
This request creates a new topic in a project with a PUT request

#### Request
`PUT /v1/projects/{project_name}/topics/{topic_name}``

#### Where
- Project_name: Name of the project to create
- Topic_name: The topic name to create

#### Example request
`curl -X PUT -H "Content-Type: application/json"  -d '' " https://{URL}/v1/projects/EGI/topics/monitoring?key=S3CR3T"`

#### Responses  

Success Response
`200 OK`
```json
{
 "name": "projects/EGI/topics/monitoring"
}
```

#### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors


# Manage Topics - Delete topic
This request deletes a topic in a project with a DELETE request

#### Request
`DELETE /v1/projects/{project_name}/topics/{topic_name}``

#### Where
- Project_name: Name of the project to delete
- Topic_name: The topic name to delete

#### Example request

`curl -X DELETE -H "Content-Type: application/json"  -d '' " https://{URL}/v1/projects/EGI/topics/monitoring?key=S3CR3T"`


#### Responses  

Success Response
Code: `200 OK`, Empty response if successful.

#### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

#Manage Topics - Get a topic
This request gets the details of a topic in a project with a GET request
#### Request
`GET /v1/projects/{project_name}/topics/{topic_name}``
#### Where
- Project_name: Name of the project to get
- Topic_name: The topic name to get

#### Example request

`curl -H "Content-Type: application/json"  -d '' " https://{URL}/v1/projects/EGI/topics/monitoring?key=S3CR3T"`

#### Responses  

Success Response
`200 OK`
```json
{
 "name": "projects/EGI/topics/monitoring"
}
```

#### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

# Manage Topics - List Topics
This request gets the list of topics in a project with a GET request
#### Request
`GET /v1/projects/{project_name}/topics`

#### Where
Project_name: Name of the project to get the list of topics

#### Example request

`curl -H "Content-Type: application/json"  -d '' " https://{URL}/v1/projects/EGI/topics/?key=S3CR3T"`

#### Responses  

Success Response
`200 OK`
```json
{
  "topics": [
    {
      "name":"/project/EGI/topics/monitoring"
    },
    {
      "name":"/project/EGI/topics/accounting"
    },
     ]
}
```

#### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

#### Publish message/s to a topic
The topic:publish endpoint publishes a message, or a list of messages to a specific topic with a  POST request

#### Request
`POST /v1/projects/{project_name}/topics/{topic_name}:publish`

#### Where
- Project_name: Name of the project to post the messages
- topic_name: to post the messages


#### Post data
```json
{
"messages": [
 	{
  		"attributes": [
   		{
    			"key": "infrastructure",
    			"value": "testing"
   		},
   		{
"key": "description",
	  		"value":"this message is used for testing purposes"
   		}
  	],
 "data":"U28geW91IHdlbnQgYWhlYWQgYW5kIGRlY29kZWQgdGhpcywgeW91IGNvdWxkbid0IHJlc2lzdCBlaCA/"

 	}
]
}
```

> The value of the data property must be always encoded in base64 format.


#### Example request

`curl -X POST -H "Content-Type: application/json"  -d { POSTDATA } https://{URL}/v1/projects/EGI/topics/monitoring:publish?key=S3CR3T"`

#### Responses  

Success Response
`200 OK`
```json
{
 "messageIds": [
  "100309303"
 ]
}
```

#### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
