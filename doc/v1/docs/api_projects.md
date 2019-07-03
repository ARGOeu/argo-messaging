#Project Api Calls

ARGO Messaging Service supports project entities as a basis of organizing and isolating groups of users & resources

## [GET] Manage Projects - List all projects
This request lists all available projects in the service

### Request
```
GET "/v1/projects"
```

### Example request
```
curl -X GET -H "Content-Type: application/json"
  "https://{URL}/v1/projects?key=S3CR3T"
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
```
curl -X GET -H "Content-Type: application/json"
  "https://{URL}/v1/projects/BRAND_NEW?key=S3CR3T"
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


```
curl -X POST -H "Content-Type: application/json"
 -d POSTDATA "https://{URL}/v1/projects/PROJECT_NEW?key=S3CR3T"
```

### Responses  
If successful, the response contains the newly created project

Success Response
`200 OK`

```json
{
 "name": "PROJET_NEW"
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
```json
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
```
curl -X PUT -H "Content-Type: application/json"
 -d POSTDATA "https://{URL}/v1/projects/PROJECT_NEW?key=S3CR3T"
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
```json
DELETE "/v1/projects/{project_name}"
```

### Where
- Project_name: Name of the project to delete

### Example request

```json
curl -X DELETE -H "Content-Type: application/json"  
 "https://{URL}/v1/projects/BRAND_NEW?key=S3CR3T"
```

### Responses  

Success Response
Code: `200 OK`, Empty response if successful.

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

```json
curl  -H "Content-Type: application/json"
"https://{URL}/v1/projects/BRAND_NEW:metrics?key=S3CR3T"
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
