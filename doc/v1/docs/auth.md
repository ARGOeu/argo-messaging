---
title: 'ARGO Messaging Service documentation | ARGO'
page_title: Authentication & Authorization
font_title: fa fa-cogs
description: Authentication & Authorization
---

# Security and privacy considerations

Authentication is the process of determining the identity of a client, which is typically a user account. Authorization is the process of determining what permissions an authenticated identity has on a set of specified resources. In the Messaging API, there can be no authorization without authentication.

> This is an initial implementation of the user authentication and authorization. In the next versions of the ARGO Messaging service we are going to provide support for both bear and OpenID Connect tokens for the API access and it will be possible to apply ACLs at each (resource) subscriptions and topics.

# User Authentication

Authentication requires the presence of a populated “users” collection in the datastore in the adhering to the following schema:

```json
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


Parameter | Description
--------- | -----------
name | username
email | User’s email
project | Project that the user belongs to
token | Secret token for authentication
roles | List of roles that user has. Each role definition is used in authorization (explained later)


Each user is authenticated by adding the url parameter ?key=T0K3N in each API request

# Authorization

Authorization requires the presence of a populated “roles” collection in the datastore in the adhering to the following schema:

```json
{
	"resource" : "resource_name:action",
	"roles" : [
		"admin",
		"member"
	]
}
```


Parameter | Description
--------- | -----------
resource | Holds the name of the resource and the action on that resource in the following format: resource_name:action
roles | A list of roles allowed on this resource:action

Resource_name:action must be the same with the default routes supported in the api currently and those are:

Action | Description
------ | -----------
topics:list | Allow user to list all topics in a project when using  `GET /projects/PROJECT_A/topics`
topics:show | Allow user to get information on a specific topic when using `GET /projects/PROJECT_A/topics/TOPIC_A`
topics:create | Allow user to create a new topic when using `PUT /projects/PROJECT_A/topics/TOPIC_NEW`
topics:delete | Allow user to delete an existing topic when using `DELETE /projects/PROJECT_A/topics/TOPIC_A`
topics:publish | Allow user to publish messages in a topic when using `POST /projects/PROJECT_A/topics/TOPIC_A:publish`
subscriptions:list | Allow user to list all subscriptions in a project when using `GET /projects/PROJECT_A/subscriptions`
subscriptions:show | Allow user to get information on a specific subscription when using `GET /projects/PROJECT_A/subscriptions/SUB_A`
subscriptions:create | Allow user to create a new subscription when using `PUT /projects/PROJECT_A/subscriptions/SUB_NEW`
subscriptions:delete | Allow user to delete an existing subscription when using `DELETE /projects/PROJECT_A/subscriptions/SUB_A`
subscriptions:pull | Allow user to pull messages from a subscription when using `POST /projects/PROJECT_A/subscriptions/SUB_A:pull`
subscriptions:acknowledge | Allow user to acknowledge messages that has pulled when using `POST /projects/PROJECT_A/subscriptions/SUB_A:acknowledge`

# Per Resource Authorization

Messaging API provides the option to control in finer detail access on resources such as topics and subscriptions for users(clients) that are producers or subscribers. Each resource (topic/subscription) comes with an access list (ACL) that contains producers or subscribers that are eligible to use that resource (when publishing or pulling messages respectively). Users with the admin role are able to modify Access lists for topics and subscriptions on the project they belong. In order for the feature to be available Messaging API should have the config parameter `per_resource_auth` set to `true`

# List ACL of a given topic

# Modify ACL of a given topic

# List ACL of a given subscription

# Modify ACL of a given subscription
