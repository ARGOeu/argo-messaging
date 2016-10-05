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


```
json

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

```
json
>>>>>>> ARGO-510 Implement API Call to create projects
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
 "name": "projects/PROJET_NEW",
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
If successful, the response contains the newly created project

Success Response
`200 OK`
```json
{
 "name": "projects/PROJET_NEW_UPDATED",
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
-d '' "https://{URL}/v1/projects/EGI?key=S3CR3T"
```

### Responses  

Success Response
Code: `200 OK`, Empty response if successful.

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
