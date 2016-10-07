#User Api Calls

ARGO Messaging Service supports calls for creating and modifing users

## [GET] Manage Users - List all users
This request lists all available users in the service

### Request
```json
GET "/v1/users"
```

### Example request
```
curl -X GET -H "Content-Type: application/json"
  "https://{URL}/v1/users?key=S3CR3T"
```

### Responses  
If successful, the response contains a list of all available users in the service

Success Response
`200 OK`

```json
{
 "users": [
    {
       "projects": [
          {
             "project_uuid": "argo_uuid",
             "roles": [
                "admin",
                "member"
             ]
          }
       ],
       "name": "Test",
       "token": "S3CR3T",
       "email": "Test@test.com",
       "service_roles":[]
    },
    {
       "projects": [
          {
             "project_uuid": "argo_uuid",
             "roles": [
                "admin",
                "member"
             ]
          }
       ],
       "name": "UserA",
       "token": "S3CR3T1",
       "email": "foo-email",
       "service_roles":[]
    },
    {
       "projects": [
          {
             "project_uuid": "argo_uuid",
             "roles": [
                "admin",
                "member"
             ]
          }
       ],
       "name": "UserB",
       "token": "S3CR3T2",
       "email": "foo-email",
       "service_roles":[]
    },
    {
       "projects": [
          {
             "project_uuid": "argo_uuid",
             "roles": [
                "consumer"
             ]
          }
       ],
       "name": "UserX",
       "token": "S3CR3T3",
       "email": "foo-email",
       "service_roles":[]
    },
    {
       "projects": [
          {
             "project_uuid": "argo_uuid",
             "roles": [
                "producer"
             ]
          }
       ],
       "name": "UserZ",
       "token": "S3CR3T4",
       "email": "foo-email",
       "service_admin":false
    }
 ]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors


## [GET] Manage Users - List a specific user
This request lists information about a specific user in the service

### Request
```
POST "/v1/users/{user_name}"
```

### Where
- user_name: Name of the user

### Example request
```
curl -X POST -H "Content-Type: application/json"
  "https://{URL}/v1/users/UserA?key=S3CR3T"
```

### Responses  
If successful, the response contains information about the specific user

Success Response
`200 OK`

```json
{
 "projects": [
    {
       "project_uuid": "argo_uuid",
       "roles": [
          "admin",
          "member"
       ]
    }
 ],
 "name": "UserA",
 "token": "S3CR3T1",
 "email": "foo-email",
 "service_roles":[]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors


## [POST] Manage Users - Create new user
This request creates a new user in a project

### Request
```json
POST "/v1/users/{user_name}"
```

### Post body:
```json
{
 "projects": [
    {
       "project_uuid": "argo_uuid",
       "roles": [
          "admin",
          "member"
       ]
    }
 ],
 "email": "foo-email",
 "service_roles":[]
}
```

### Where
- user_name: Name of the user
- projects: A list of Projects & associated roles that the user has on those projects
- email: User's email
- service_roles: A list of service-wide roles. An example of service-wide role is `service_admin` which can manage projects or other users

### Example request
```
json
curl -X POST -H "Content-Type: application/json"
 -d POSTDATA "https://{URL}/v1/projects/ARGO/users/USERNEW?key=S3CR3T"
```

### Responses  
If successful, the response contains the newly created project

Success Response
`200 OK`
```json
{
 "projects": [
    {
       "project_uuid": "argo_uuid",
       "roles": [
          "admin",
          "member"
       ]
    }
 ],
 "name": "USERNEW",
 "token": "R4ND0MT0K3N",
 "email": "foo-email",
 "service_admin":false
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
