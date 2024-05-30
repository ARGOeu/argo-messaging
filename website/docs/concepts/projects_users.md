---
id: projects_users
title: Initial Project & User Management
sidebar_position: 5
---

This document describes some of the more advanced setup you may need to do while configuring and deploying the ARGO
Messaging Service.

## A typical quick-start scenario

After a fresh install of the ARGO Messaging Service, the steps you need to follow are:

- Configure `service_token`: to enable the service.
- Create a service_admin user: to start managing the service.
- Create a project: Project entities is used as a basis of organizing and isolating groups of users & resources
- Create a project_admin user: Users that have the project_admin have, by default, all capabilities in their project.
  They can also manage resources such as topics and subscriptions (CRUD) and also manage ACLs (users) on those resources
  as well.
- Create a topic: The main resource that is scoped in a project, and can hold messages.
- Create a subscription: A subscription is the main resource from which users consume messages.
- Create users for the new resources: Usually a project has publisher and consumer accounts for clients that either are
  authorized to publish or consume messages.

## Configure `service_token`

ARGO Messaging Service configuration includes the `service_token` parameter. This `service_token` configuration
parameter can be used to create the first `service_admin` user of the service

First a service token must be defined in the config.json as such:

```
{
  "bind_ip":"",
  "port":8080,
  "zookeeper_hosts":["localhost"],
  "kafka_znode":"",
  "store_host":"localhost",
  "store_db":"argo_msg",
  "certificate":"/etc/pki/tls/certs/localhost.crt",
  "certificate_key":"/etc/pki/tls/private/localhost.key",
  "per_resource_auth":"true",
  "service_token":"S3CR3T",
  "push_enabled": false
}
```

The service token in this example has the value: `S3CR3T`
This `service_token` is authorized for all available actions (projects,users,topics,subscriptions).

In order to enable the use of this `service_token` you must restart the service.

```
service argo-messaging restart
```

## Create a service_admin user

The service_token is intended to be used for the first initialization of the API. The first thing the service needs is a
user with all possible capabilities, which is a `service_admin`.
Now even though no user has been initialized in the service, the administrator can use the ARGO Messaging API Call with
service_token `S3CR3T` to create the user.
The service_admin will be able to further define projects and other users.

Using the service_token an admin can create a new service_admin user with the username `demo_service_admin` by calling:

```
POST https://{URL}/v1/users/demo_service_admin
```

with the following POST BODY:

```json
{
  "email": "sadmin@mail.example.foo",
  "service_roles": [
    "service_admin"
  ]
}
```

It is important to mention that the user has the **"service_admin"** role defined in the service_roles list in order to
be a service_admin.

The response:

```json
{
  "projects": [],
  "name": "demo_service_admin",
  "token": "904c56cc6e2b1955dbd98ace80a45be8238432fc",
  "email": "sadmin@mail.example.foo",
  "service_roles": [
    "service_admin"
  ],
  "created_on": "2016-10-13T11:19:07Z",
  "modified_on": "2016-10-13T11:19:07Z"
}
```

The generated token `904c56cc6e2b1955dbd98ace80a45be8238432fc` can be used to authenticate the new user.
For more details visit the [API Users](/api_advanced/api_users.md) to see all possible actions for users.

## Create a project

Using the `demo_service_admin` account, the user can create the first project (ex named 'DEMO') by issuing:

```
POST https://{URL}/v1/projects/DEMO
```

with the following POST BODY:

```json
{
  "description": "my first demo project"
}
```

and the response:

```json
{
  "name": "DEMO",
  "created_on": "2016-10-13T12:19:07Z",
  "modified_on": "2016-10-13T12:19:07Z",
  "created_by": "demo_service_admin",
  "description": "my first demo project"
}
```

Response informs that the project has been indeed `created_by` the `demo_service_admin` user.

For more details visit the [API Projects](/api_advanced/api_projects.md) to see all possible actions for projects.

## Create a project_admin

Service_admin users are not attached to specific projects. Instead each project should have a `project_admin` user that
will manage topics, subscriptions and ACLs on those resources. To create a `project_admin` user in project `DEMO`, the
user `demo_service_admin` will issue:

```
POST https://{URL}/v1/users/admin_DEMO
```

with the following POST BODY:

```json
{
  "email": "demoadmin@mail.example.foo",
  "projects": [
    {
      "project": "DEMO",
      "roles": [
        "project_admin"
      ]
    }
  ]
}
```

The user definition (in POST body) should have the field `projects` defined. The field accepts a list of tuple items (
project,roles) which describe each project that the user participates to and under which roles. A user can have multiple
roles in a project and also participate in multiple projects as well. In this example, the user must participate in
project `DEMO` and under the role of `project_admin`.

The response:

```json
{
  "projects": [
    {
      "project": "DEMO",
      "roles": [
        "project_admin"
      ]
    }
  ],
  "name": "admin_DEMO",
  "token": "6311196665befcc1523b8e013979347b8780254c",
  "email": "demoadmin@mail.example.foo",
  "service_roles": [],
  "created_on": "2016-10-13T12:29:07Z",
  "modified_on": "2016-10-13T12:29:07Z",
  "created_by": "demo_service_admin"
}
```

For more details visit the [API Users](/api_advanced/api_users.md) to see all possible actions for users.

## Create a topic

Service_admin users don't manage resources such as topics/subscriptions. Instead in each project the project_admin is
eligible for creating (and managing) topics and subscriptions. To create a new topic (named `topic101`) as `admin_DEMO`
user in project `DEMO` the user issues:

```
PUT https://{URL}/v1/projects/DEMO/topics/topic101
```

with response:

```
{
  "name": "/projects/DEMO/topics/topic101"
}
```

Notice that the token used in api `key` changes to that of the `admin_DEMO` user

For more details visit the [API Topics](/api_advanced/api_topics.md) to see all possible actions for topics.

## Create a subscription

To create a new subscription (named `sub101`) to topic `topic101` of project `DEMO` the `admin_DEMO` user issues:

```
PUT https://{URL}/v1/projects/DEMO/subscriptions/subs101
```

with POST Body:

```json
{
  "topic": "projects/DEMO/topic/topic101"
}
```

and response:

```json
{
  "name": "/projects/DEMO/subscriptions/sub101",
  "topic": "/projects/DEMO/topics/topic101",
  "pushConfig": {
    "pushEndpoint": "",
    "retryPolicy": {}
  },
  "ackDeadlineSeconds": 10
}
```

For more details visit the [API Subscriptions](/api_advanced/api_subs.md) to see all possible actions for Subscriptions.

## Create users for the new resources

Usually a project will have also publisher and consumer accounts for clients that either are authorized to publish or
consume messages. The user `demo_service_admin` can create a `publisher_DEMO` and `consumer_DEMO` for project `DEMO` as
such:

To create the `publisher_DEMO` user:

```
POST https://{URL}/v1/users/publisher_DEMO
```

with POST Body:

```json
{
  "email": "demopublisher@mail.example.foo",
  "projects": [
    {
      "project": "DEMO",
      "roles": [
        "publisher"
      ]
    }
  ]
}
```

resulting in response:

```json
{
  "projects": [
    {
      "project": "DEMO",
      "roles": [
        "publisher"
      ]
    }
  ],
  "name": "publisher_DEMO",
  "token": "915dff62846dd1d790b4296c034c184fa3a859b6",
  "email": "demopublisher@mail.example.foo",
  "service_roles": [],
  "created_on": "2016-10-13T12:39:07Z",
  "modified_on": "2016-10-13T12:39:07Z",
  "created_by": "demo_service_admin"
}
```

To create the `consumer_DEMO` user:

```
POST https://{URL}/v1/users/consumer_DEMO
```

with POST Body:

```json
{
  "email": "democonsumer@mail.example.foo",
  "projects": [
    {
      "project": "DEMO",
      "roles": [
        "consumer"
      ]
    }
  ]
}
```

resulting in response:

```json
{
  "projects": [
    {
      "project": "DEMO",
      "roles": [
        "consumer"
      ]
    }
  ],
  "name": "consumer_DEMO",
  "token": "dba38fd1a45337a617a59e7278c756f23642e9e7",
  "email": "democonsumer@mail.example.foo",
  "service_roles": [],
  "created_on": "2016-10-13T12:40:07Z",
  "modified_on": "2016-10-13T12:40:07Z",
  "created_by": "demo_service_admin"
}
```

For more details visit the [API Users](/api_advanced/api_users.md) to see all possible actions for users.

### Modify topic ACL to give access to publisher

In order to give access to user `publisher_DEMO` to topic `topic101`, the user `admin_DEMO` must modify the topic's ACL
as such:

```
POST https://{URL}/v1/projects/DEMO/topics/topic101:modifyAcl
```

with POST body:

```json
{
  "authorized_users": [
    "publisher_DEMO"
  ]
}
```

and empty response with `200 OK`

Now the user `publisher_DEMO` will be authorized to call action `topic:publish` on `topic101` and send messages

### Modify subscription ACL to give access to consumer

In order to give access to user `consumer_DEMO` to subscription `sub101`, the user `admin_DEMO` must modify the
subscription's ACL as such:

```
POST https://{URL}/v1/projects/DEMO/subscriptions/sub101:modifyAcl
```

with POST body:

```json
{
  "authorized_users": [
    "consumer_DEMO"
  ]
}
```

and empty response with `200 OK`

Now the user `consumer_DEMO` will be authorized to call action `subscription:pull` on `sub101` and consume messages
