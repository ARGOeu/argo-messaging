#Project Api Calls

ARGO Messaging Service supports project entities as a basis of organizing and isolating groups of users & resources

## [GET] Manage Projects - List all projects
This request lists all available projects in the service

### Request
```
GET "/v1/projects"
```

### Example request
```bash
curl -X GET -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
  "https://{URL}/v1/projects"
```

### Responses  
If successful, the response contains a list of all available projects

Success Response
`200 OK`


```json
{
 "projects": [
    {
       "name": "PROJECT_1",
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z",
       "created_by": "userA",
       "description": "simple project"
    },
    {
       "name": "PROJECT_2",
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z",
       "created_by": "userA",
       "description": "simple project"
    },
    {
       "name": "BRAND_NEW",
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z",
       "created_by": "userA",
       "description": "brand new project"
    }
 ]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors


## [GET] Manage Projects - List a specific project
This request lists information about a specific project

### Request
```
GET "/v1/projects/{project_name}"
```

### Where
- Project_name: Name of the project to get information on


### Example request
```bash
curl -X GET -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
  "https://{URL}/v1/projects/BRAND_NEW"
```

### Responses  
If successful, the response contains information about the specific project

Success Response
`200 OK`

```json
{
   "name": "BRAND_NEW",
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z",
   "created_by": "userA",
   "description": "brand new project"
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [POST] Manage Projects - Create new project
This request creates a new project with the given project_name with a POST request

### Request

```
POST "/v1/projects/{project_name}"
```

### Where
- Project_name: Name of the project to create

### Post body:

```json
{
  "description" : "a simple description"
}
```

### Example request


```bash
curl -X POST -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
 -d $POSTDATA "https://{URL}/v1/projects/PROJECT_NEW"
```

### Responses  
If successful, the response contains the newly created project

Success Response
`200 OK`

```json
{
 "name": "PROJECT_NEW"
 "created_on": "2009-11-10T23:00:00Z",
 "modified_on": "2009-11-10T23:00:00Z",
 "created_by": "userA",
 "description": "brand new project"
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [PUT] Manage Projects - Update a project
This request updates information (such as name,description) on an existing project (PUT)

### Request
```
PUT "/v1/projects/{project_name}"
```

### Where
- Project_name: Name of the project to create

### PUT body:
```json
{
  "name":"new project name",
  "description" : "a simple description"
}
```

### Example request
```bash
curl -X PUT -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
 -d $POSTDATA "https://{URL}/v1/projects/PROJECT_NEW"
```

### Responses  
If successful, the response contains the newly updated

Success Response
`200 OK`
```json
{
 "name": "PROJET_NEW_UPDATED",
 "created_on": "2009-11-10T23:00:00Z",
 "modified_on": "2009-11-13T23:00:00Z",
 "created_by": "userA",
 "description": "description project updated"
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors


## [DELETE] Manage Projects - Delete Project
This request deletes a specific project

### Request
```
DELETE "/v1/projects/{project_name}"
```

### Where
- Project_name: Name of the project to delete

### Example request

```bash
curl -X DELETE -H "Content-Type: application/json"  
 "https://{URL}/v1/projects/BRAND_NEW?key=S3CR3T"
```

### Responses  

Success Response
Code: `200 OK`, Empty response if successful.

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

### [GET] List all users that are members of a specific project

- `details`, if set to `true`, it will return the detailed view of each user,
containing the projects, subscriptions and topics that the user belongs to.


### Example request
```bash
curl -X GET -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
  "https://{URL}/v1/projects/ARGO2/members?details=true"
```

### Responses  
If successful, the response contains a list of all available users in the specific project

Success Response
`200 OK`

```json
{
 "users": [
    {
       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebw",
       "projects": [
          {
             "project": "ARGO2",
             "roles": [
                "consumer",
                "publisher"
             ],
             "topics": [],
             "subscriptions": []
          }
       ],
       "name": "Test",
       "token": "S3CR3T",
       "email": "Test@test.com",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z"
    }
 ],
 "nextPageToken": "",
 "totalSize": 1
}
```

### The Unprivileged mode (non service_admin user)
When a user is project_admin instead of service_admin and lists a project's users the results
returned remove user information such as `token`, `service_roles` and `created_by` For example:

```json
{
 "users": [
    {
       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebw",
       "projects": [
          {
             "project": "ARGO2",
             "roles": [
                "consumer",
                "publisher"
             ],
             "topics": [],
             "subscriptions": []
          }
       ],
       "name": "Test",
       "token": "",
       "email": "Test@test.com",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z"
    }
 ],
 "nextPageToken": "",
 "totalSize": 1
}
```

### [GET] Show a specific member user of the specific project

### Example request
```bash
curl -X GET -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
  "https://{URL}/v1/projects/ARGO2/members/Test"
```

### Responses  
If successful, the response contains information of the specific user Test

Success Response
`200 OK`

```json
{
 "users": [
    {
       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebw",
       "projects": [
          {
             "project": "ARGO2",
             "roles": [
                "consumer",
                "publisher"
             ],
             "topics": [],
             "subscriptions": []
          }
       ],
       "name": "Test",
       "token": "S3CR3T",
       "email": "Test@test.com",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z"
    }
 ],
 "nextPageToken": "",
 "totalSize": 1
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

### [POST] Create a new member user under the specific project

### Example request
```bash
curl -X POST -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
  -d $POSTDATA "https://{URL}/v1/projects/ARGO2/members/NewUser"
```

### Post body:

```json
{
	"projects": [
			{
				"project": "ARGO2",
				"roles": ["consumer"]
			}
		],
	"email": "email@test.com"
}
```

### Responses  
If successful, the response contains information about the newly created user

Success Response
`200 OK`

```json
{

       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebw",
       "projects": [
          {
             "project": "ARGO2",
             "roles": [
                "consumer"
             ],
             "topics": [],
             "subscriptions": []
          }
       ],
       "name": "NewUSer",
       "token": "S3CR3T",
       "email": "email@test.com",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z"
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

### [PUT] Updates the roles for a member user under the specific project

### Example request
```bash
curl -X PUT -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
  -d $POSTDATA "https://{URL}/v1/projects/ARGO2/members/NewUser"
```

### Post body:

```json
{
	"projects": [
			{
				"project": "ARGO2",
				"roles": ["consumer"]
			}
		]
}
```

### Responses  
If successful, the response contains information about the updated user

Success Response
`200 OK`

```json
{

       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebw",
       "projects": [
          {
             "project": "ARGO2",
             "roles": [
                "consumer"
             ],
             "topics": [],
             "subscriptions": []
          }
       ],
       "name": "NewUSer",
       "token": "S3CR3T",
       "email": "email@test.com",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z"
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

### [POST] Add/Invite a user to a project

### Example request
```bash
curl -X POST -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
  -d $POSTDATA "https://{URL}/v1/projects/ARGO2/members/NewUser:add"
```

### Post body:

```json
{
  "roles": ["consumer"]
}
```

### Responses  
If successful, the response contains information about the added user

Success Response
`200 OK`

```json
{

       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebw",
       "projects": [
          {
             "project": "ARGO2",
             "roles": [
                "consumer"
             ],
             "topics": [],
             "subscriptions": []
          }
       ],
       "name": "NewUSer",
       "token": "S3CR3T",
       "email": "email@test.com",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z"
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors


### [POST] Remove a user from the project

### Example request
```bash
curl -X POST -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
 "https://{URL}/v1/projects/ARGO2/members/NewUser:remove"
```

### Responses  
Empty response on success
`200 OK`

```json
{}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors


## [GET] Project Metrics
The following request returns related metrics for the specific project: eg. the number of topics

### Request
```
GET "/v1/projects/{project_name}:metrics"
```

### Where
- Project_name: name of the project
- topic_name: name of the topic

### Example request

```bash
curl  -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
"https://{URL}/v1/projects/BRAND_NEW:metrics"
```



### Responses  
If successful it returns projects related metrics (number of topics, number of subscriptions).

Success Response
`200 OK`
```
{
   "metrics": [
      {
         "metric": "project.number_of_topics",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project",
         "resource_name": "ARGO",
         "timeseries": [
            {
               "timestamp": "2017-06-30T13:53:13Z",
               "value": 3
            }
         ],
         "description": "Counter that displays the number of topics belonging to the specific project"
      },
      {
         "metric": "project.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project",
         "resource_name": "ARGO",
         "timeseries": [
            {
               "timestamp": "2017-06-30T13:53:13Z",
               "value": 4
            }
         ],
         "description": "Counter that displays the number of subscriptions belonging to the specific project"
      },
      {
         "metric": "project.user.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserA",
         "timeseries": [
            {
               "timestamp": "2017-06-30T13:53:13Z",
               "value": 3
            }
         ],
         "description": "Counter that displays the number of subscriptions that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserB",
         "timeseries": [
            {
               "timestamp": "2017-06-30T13:53:13Z",
               "value": 3
            }
         ],
         "description": "Counter that displays the number of subscriptions that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserX",
         "timeseries": [
            {
               "timestamp": "2017-06-30T13:53:13Z",
               "value": 1
            }
         ],
         "description": "Counter that displays the number of subscriptions that a user has access to the specific project"
      },
      {
         "metric": "project.user.number_of_subscriptions",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project.user",
         "resource_name": "ARGO.UserZ",
         "timeseries": [
            {
               "timestamp": "2017-06-30T13:53:13Z",
               "value": 2
            }
         ],
         "description": "Counter that displays the number of subscriptions that a user has access to the specific project"
      },
      {
         "metric": "project.number_of_daily_messages",
         "metric_type": "counter",
         "value_type": "int64",
         "resource_type": "project",
         "resource_name": "ARGO",
         "timeseries": [
            {
               "timestamp": "2018-10-02",
               "value": 30
            },
            {
               "timestamp": "2018-10-01",
               "value": 110
            }
         ],
         "description": "A collection of counters that represents the total number of messages published each day to all of the project's topics"
      }
   ]
}

```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
