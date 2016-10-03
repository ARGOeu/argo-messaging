#Project Api Calls

ARGO Messaging Service supports project entities as a basis of organizing and isolating groups of users & resources

## [GET] Manage Projects - List all projects
This request lists all avaliable projects in the service

### Request
```json
GET "/v1/projects"
```

### Example request
```json
curl -X GET -H "Content-Type: application/json"
  "https://{URL}/v1/projects?key=S3CR3T"
```

### Responses  
If successful, the response contains a list of all available projects

Success Response
`200 OK`
```
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


## [GET] Manage Projects - List a specific project
This request lists information about a specific project

### Request
```json
POST "/v1/projects/{project_name}"
```

### Where
- Project_name: Name of the project to get information on


### Example request
```json
curl -X POST -H "Content-Type: application/json"
  "https://{URL}/v1/projects/PROJECT_NEW?key=S3CR3T"
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


## [POST] Manage Projects - Create new project
This request creates a new project with the given project_name with a POST request

### Request
```json
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
```json
curl -X POST -H "Content-Type: application/json"
 -d POSTDATA "https://{URL}/v1/projects/PROJECT_NEW?key=S3CR3T"
```

### Responses  
If successful, the response contains the newly created project

Success Response
`200 OK`
```json
{
 "name": "projects/PROJET_NEW"
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
