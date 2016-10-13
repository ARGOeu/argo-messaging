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
            "project": "ARGO",
             "roles": [
                "project_admin",
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
            "project": "ARGO",
             "roles": [
                "project_admin",
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
            "project": "ARGO",
             "roles": [
                "project_admin",
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
            "project": "ARGO",
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
             "project": "ARGO",
             "roles": [
                "publisher"
             ]
          }
       ],
       "name": "UserZ",
       "token": "S3CR3T4",
       "email": "foo-email",
       "service_roles":[]
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
GET "/v1/users/{user_name}"
```

### Where
- user_name: Name of the user

### Example request
```
curl -X GET -H "Content-Type: application/json"
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
       "project_uuid": "ARGO",
       "roles": [
          "project_admin",
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
       "project": "ARGO",
       "roles": [
          "project_admin",
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


##### Available Roles
ARGO Messaging Service has the following predefined project roles:

| Role | Description |
|------|-------------|
| project_admin  | Users that have the `project_admin` have, by default, all capabilities in their project. They can also manage resources such as topics and subscriptions (CRUD) and also manage ACLs (users) on those resources as well |
| consumer | Users that have the `consumer` role are only able to pull messages from subscriptions that are authorized to use (based on ACLs)
| publisher | Users that have the `publisher` role are only able to publish messages on topics that are authorized to use (based on ACLs)

and the following service-wide role:

| Role | Description |
|------|-------------|
| service_admin  | Users with `service_admin` role operate service wide. They are able to create, modify and delete projects. Also they are able to create, modify and delete users and assign them to projects.  

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
       "project": "ARGO",
       "roles": [
          "project_admin",
       ]
    }
 ],
 "name": "USERNEW",
 "token": "R4ND0MT0K3N",
 "email": "foo-email",
 "service_roles":[]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [PUT] Manage Users - Update a user
This request updates an existing user's information

### Request
```json
PUT "/v1/users/{user_name}"
```

### Put body:
```json
{
 "projects": [
    {
       "project": "ARGO2",
       "roles": [
          "project_admin",
       ]
    }
 ],
 "name": "CHANGED_NAME",
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
 -d PUTDATA "https://{URL}/v1/projects/ARGO/users/USERNEW?key=S3CR3T"
```

### Responses  
If successful, the response contains the newly created project

Success Response
`200 OK`

```json
{
 "projects": [
    {
       "project": "ARGO2",
       "roles": [
          "project_admin",
       ]
    }
 ],
 "name": "CHANGED_NAME",
 "token": "R4ND0MT0K3N",
 "email": "foo-email",
 "service_roles":[]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [POST] Manage Users - Refresh token
This request refreshes an existing user's token
### Request

```json
POST "/v1/users/{user_name}:refreshToken"
```
### Where
- user_name: Name of the user


### Example request
```
json
curl -X POST -H "Content-Type: application/json"
 "https://{URL}/v1/projects/ARGO/users/USER2:refreshToken?key=S3CR3T"
```

### Responses  
If successful, the response contains the newly created project

Success Response
`200 OK`

```json
{
 "projects": [
    {
       "project": "ARGO",
       "roles": [
          "project_admin",
       ]
    }
 ],
 "name": "USER2",
 "token": "NEWRANDOMTOKEN",
 "email": "foo-email",
 "service_roles":[]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [DELETE] Manage Users - Delete User
This request deletes an existing user
### Request

```json
DELETE "/v1/users/{user_name}"
```

### Where
- user_name: Name of the user


### Example request
``` json
curl -X DELETE -H "Content-Type: application/json"
 "https://{URL}/v1/projects/ARGO/users/USER2?key=S3CR3T"
```

### Responses  
If successful, the response returns empty

Success Response
`200 OK`


### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
