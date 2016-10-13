# Initial Project and user management

ARGO Messaging Service configuration includes the `service_token` parameter. Administrators can use this
feature to define and use an access token for initial project and user configuration.

## Available Roles
ARGO Messaging Service has the following predefined project roles:

| Role | Description |
|------|-------------|
| project_admin  | Users that have the `project_admin` role are assigned to projects which are able to modify or delete. Also they are able to manage resources such as topics and subscriptions (CRUD) and also manage ACLs on those resources as well |
| consumer | Users that have the `consumer` role are only able to pull messages from subscriptions that are authorized to use (based on ACLs)
| publisher | Users that have the `publisher` role are only able to publish messages on topics that are authorized to use (based on ACLs)

and the following service-wide role:

| Role | Description |
|------|-------------|
| service_admin  | Users with `service_admin` role operate service wide. They are able to create, modify and delete projects. Also they are able to create, modify and delete users and assign them to projects.  


## A typical quick-start scenario

After a fresh install of the argo-messaging , the `service_token` configuration parameter can be used to create the first `service_admin` user of the service

First a service token must be defined in the config.json as such:
{
  "bind_ip":"",
  "port":8080,
  "zookeeper_hosts":["localhost"],
  "store_host":"localhost",
  "store_db":"argo_msg",
  "certificate":"/etc/pki/tls/certs/localhost.crt",
  "certificate_key":"/etc/pki/tls/private/localhost.key",
  "per_resource_auth":"true",
  "service_token":"S3CR3T"
}

The service token in this example has the value: `S3CR3T`

Now even though no user has been initialized in the service, the administrator can use the ARGO-Messaging api with token `S3CR3T` as an api key in each request. The token is authorized for all available actions (projects,users,topics,subscriptions).

### Generate a service_admin user

The service_token is intended to be used for the first initialization of the API. Then a user with a service_admin role can be defined. The service_admin will be able to further define projects and other users.

Using the service_token an admin can create a new service_admin user by calling:

```
POST https://{URL}/v1/users/demo_service_admin?key=S3CR3T
```
with the following POST BODY:
```json
{
   "email":"sadmin@mail.example.foo",
   "service_roles":["service_admin"]
}
```

It is important that the user has the "service_admin" role defined in the service_roles list in order to be a service_admin.

The response:
```json
{
  "projects": [],
  "name": "demo_service_admin",
  "token": "904c56cc6e2b1955dbd98ace80a45be8238432fc",
  "email": "sadmin@mail.example.foo",
  "service_roles": [
    "service_admin"
  ]
}
```

The generated token `904c56cc6e2b1955dbd98ace80a45be8238432fc` can be used to authenticate as the user `demo_service_admin` and create the first project.

### Create the first project named DEMO

Using the `demo_service_admin` account, the user will create the first project (named 'DEMO') by issuing:

```
POST https://{URL}/v1/projects/DEMO?key=904c56cc6e2b1955dbd98ace80a45be8238432fc
```
with the following POST BODY:
```
{
   "description":"my first demo project"
}
```

and the response:
```
{
  "name": "DEMO",
  "created_on": "2016-10-13T12:19:07.341+03:00",
  "modified_on": "2016-10-13T12:19:07.341+03:00",
  "created_by": "demo_service_admin",
  "description": "my first demo project"
}
```

Response informs that the project has been indeed `created_by` the `demo_service_admin` user

### Create a project_admin user in project DEMO

Service_admin users are not attached to specific projects. Instead each project should have a `project_admin` user that will manage topics, subscriptions and ACLs on those resources. To create a `project_admin` user in project `DEMO`, the user `demo_service_admin` will issue:

```
POST https://{URL}/v1/users/admin_DEMO?key=904c56cc6e2b1955dbd98ace80a45be8238432fc
```
with the following POST BODY:
```json
{
   "email":"demoadmin@mail.example.foo",
   "projects":[{"project":"DEMO","roles":["project_admin"]}]
}
```
The user definition (in POST body) should have the field `projects` defined. The field accepts a list of tuple items (project,roles) which describe each project that the user participates to and under which roles. A user can have multiple roles in a project and also participate in multiple projects as well. In this example, the user must participate in project `DEMO` and under the role of `project_admin`.

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
  "service_roles": []
}
```

### Create a consumer user and a publisher user in project DEMO

Usually a project will have also publisher and consumer accounts for clients that either are authorized to publish or consume messages. The user `demo_service_admin` can create a `publisher_DEMO` and `consumer_DEMO` for project `DEMO` as such:

To create the `publisher_DEMO` user:

```
POST https://{URL}/v1/users/publisher_DEMO?key=904c56cc6e2b1955dbd98ace80a45be8238432fc
```
with POST Body:
```json
{
   "email":"demopublisher@mail.example.foo",
   "projects":[{"project":"DEMO","roles":["publisher"]}]
}
```

resulting in response:
```
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
  "service_roles": []
}
```

To create the `conumer_DEMO` user:
```
POST https://{URL}/v1/users/consumer_DEMO?key=904c56cc6e2b1955dbd98ace80a45be8238432fc
```
with POST Body:
```json
{
   "email":"democonsumer@mail.example.foo",
   "projects":[{"project":"DEMO","roles":["consumer"]}]
}

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
  "service_roles": []
}
```

### Create a topic in project DEMO as project_admin

Service_admin users don't manage resources such as topics/subscriptions. Instead in each project the project_admin is eligible for creating (and managing) topics and subscriptions. To create a new topic (named `topic101`) as `admin_DEMO` user in project `DEMO` the user issues:

```
PUT https://{URL}/v1/projects/DEMO/topics/topic101?key=6311196665befcc1523b8e013979347b8780254c
```

with response:
```
{
  "name": "/projects/DEMO/topics/topic101"
}
```

Notice that the token used in api `key` changes to that of the `admin_DEMO` user

### Create a subscription in project DEMO as project_admin

To create a new subscription (named `sub101`) to topic `topic101` of project `DEMO` the `admin_DEMO` user issues:

```
PUT https://{URL}/v1/projects/DEMO/subscriptions/subs101?key=6311196665befcc1523b8e013979347b8780254c
```
with POST Body:
```
{
   "topic":"projects/DEMO/topic/topic101"
}
```

and response:
```
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

### Modify topic ACL to give access to publisher

In order to give access to user `publisher_DEMO` to topic `topic101`, the user `admin_DEMO` must modify the topic's ACL as such:

```
POST https://{URL}/v1/projects/DEMO/topics/topic101:modifyAcl?key=6311196665befcc1523b8e013979347b8780254c
```
with POST body:
```json
{
   "authorized_users":["publisher_DEMO"]
}
```
and empty response with `200 OK`

Now the user `publisher_DEMO` will be authorized to call action `topic:publish` on `topic101` and send messages

### Modify subscription ACL to give access to consumer

In order to give access to user `consumer_DEMO` to subscription `sub101`, the user `admin_DEMO` must modify the subscription's  ACL as such:

```
POST https://{URL}/v1/projects/DEMO/subscriptions/sub101:modifyAcl?key=6311196665befcc1523b8e013979347b8780254c
```
with POST body:
```json
{
   "authorized_users":["consumer_DEMO"]
}
```
and empty response with `200 OK`

Now the user `consumer_DEMO` will be authorized to call action `subscription:pull` on `sub101` and consume messages
