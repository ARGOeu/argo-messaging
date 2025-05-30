---
id: auth
title: Authentication & Authorization
sidebar_position: 4
---

# Security and privacy considerations

Authentication is the process of determining the identity of a client, which is typically a user account. Authorization
is the process of determining what permissions an authenticated identity has on a set of specified resources. In the
Messaging API, there can be no authorization without authentication.

> This is an initial implementation of the user authentication and authorization. In the next versions of the ARGO
> Messaging service we are going to provide support for both bear and OpenID Connect tokens for the API access and it will
> be possible to apply ACLs at each (resource) subscriptions and topics.

## User Authentication

Authentication requires the presence of a populated “users” collection in the datastore in the adhering to the following
schema:

```
{
	"name" : "john",
	"email" : "john@doe.com",
	"project" : "ARGO",
	"token" : "S3CR3T",
	"roles" : [
		"admin",
		"member"
	]
}
```

| Parameter | Description                                                                                  |
|-----------|----------------------------------------------------------------------------------------------|
| name      | username                                                                                     |
| email     | User’s email                                                                                 |
| project   | Project that the user belongs to                                                             |
| token     | Secret token for authentication                                                              |
| roles     | List of roles that user has. Each role definition is used in authorization (explained later) |

Each user is authenticated by adding the header parameter `x-api-key` in each API request

## Authorization

Authorization requires the presence of a populated “roles” collection in the datastore in the adhering to the following
schema:

```
{
	"resource" : "resource_name:action",
	"roles" : [
		"admin",
		"member"
	]
}
```

| Parameter | Description                                                                                                  |
|-----------|--------------------------------------------------------------------------------------------------------------|
| resource  | Holds the name of the resource and the action on that resource in the following format: resource_name:action |
| roles     | A list of roles allowed on this resource:action                                                              |

Resource_name:action must be the same with the default routes supported in the api currently and those are:

| Action                    | Description                                                                                                              |
|---------------------------|--------------------------------------------------------------------------------------------------------------------------|
| topics:list               | Allow user to list all topics in a project when using  `GET /projects/PROJECT_A/topics`                                  |
| topics:show               | Allow user to get information on a specific topic when using `GET /projects/PROJECT_A/topics/TOPIC_A`                    |
| topics:create             | Allow user to create a new topic when using `PUT /projects/PROJECT_A/topics/TOPIC_NEW`                                   |
| topics:delete             | Allow user to delete an existing topic when using `DELETE /projects/PROJECT_A/topics/TOPIC_A`                            |
| topics:publish            | Allow user to publish messages in a topic when using `POST /projects/PROJECT_A/topics/TOPIC_A:publish`                   |
| subscriptions:list        | Allow user to list all subscriptions in a project when using `GET /projects/PROJECT_A/subscriptions`                     |
| subscriptions:show        | Allow user to get information on a specific subscription when using `GET /projects/PROJECT_A/subscriptions/SUB_A`        |
| subscriptions:create      | Allow user to create a new subscription when using `PUT /projects/PROJECT_A/subscriptions/SUB_NEW`                       |
| subscriptions:delete      | Allow user to delete an existing subscription when using `DELETE /projects/PROJECT_A/subscriptions/SUB_A`                |
| subscriptions:pull        | Allow user to pull messages from a subscription when using `POST /projects/PROJECT_A/subscriptions/SUB_A:pull`           |
| subscriptions:acknowledge | Allow user to acknowledge messages that has pulled when using `POST /projects/PROJECT_A/subscriptions/SUB_A:acknowledge` |

## Per Resource Authorization

Messaging API provides the option to control in finer detail access on resources such as topics and subscriptions for
users(clients) that are producers or subscribers. Each resource (topic/subscription) comes with an access list (ACL)
that contains producers or subscribers that are eligible to use that resource (when publishing or pulling messages
respectively). Users with the admin role are able to modify Access lists for topics and subscriptions on the project
they belong. In order for the feature to be available Messaging API should have the config parameter `per_resource_auth`
set to `true`

## [GET] List ACL of a given topic

Please refer to section [Topics:List ACL of a given topic ](/api_advanced/api_topics.md#get-list-acl-of-a-given-topic).

## [POST] Modify ACL of a given topic

Please refer to
section [Topics:Modify ACL of a given topic ](/api_advanced/api_topics.md#post-modify-acl-of-a-given-topic).

## [GET] List ACL of a given subscription

The following request returns a list of authorized users for a given subscription

### Request

`GET /v1/projects/{project_name}/subscriptions/{sub_name}:acl`

### Where

- Project_name: Name of the project
- sub_name: name of the subscription

### Example request

```bash
curl -X POST -H "Content-Type: application/json" -H "x-api-key: S3CR3T"  
-d $POSTDATA "https://{URL}/v1/projects/EGI/subscriptions/monitoring:acl"
```

### Responses

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

Please refer to section [Errors](/api_basic/api_errors.md) to see all possible Errors

## [POST] Modify ACL of a given subscription

The following request Modifies the authorized users list of a given subscription

### Request

`POST /v1/projects/{project_name}/subscriptions/{sub_name}:modifyAcl`

### Where

- Project_name: Name of the project
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
curl -H "Content-Type: application/json" -H "x-api-key: S3CR3T" 
 "https://{URL}/v1/projects/EGI/subscriptions/monitoring:modifyAcl"
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

Please refer to section [Errors](/api_basic/api_errors.md) to see all possible Errors
