#User Api Calls

ARGO Messaging Service supports calls for creating and modifing users

## [GET] Manage Users - List all users
This request lists all available users in the service using pagination

It is important to note that if there are no results to return the service will return the following:

Success Response
`200 OK`

```json
{
 "users": [],
  "nextPageToken": "",
  "totalSize": 0
 }
```
Also the default value for `pageSize = 0` and `pageToken = "`.

`Pagesize = 0` returns all the results.
### Request
```json
GET "/v1/users"
```

### Paginated Request that returns all users in one page

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
    },
    {
       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",
       "projects": [
          {
             "project": "ARGO",
             "roles": [
                "consumer",
                "publisher"
             ],
             "topics": [
                "topic1",
                "topic2"
             ],
             "subscriptions": [
                "sub1",
                "sub2",
                "sub3"
             ]
          }
       ],
       "name": "UserA",
       "first_name": "FirstA",
       "last_name": "LastA",
       "organization": "OrgA",
       "description": "DescA",
       "token": "S3CR3T1",
       "email": "foo-email",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z"
    },
    {
       "uuid": "94bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",
       "projects": [
          {
             "project": "ARGO",
             "roles": [
                "consumer",
                "publisher"
             ],
             "topics": [
                "topic1",
                "topic2"
             ],
             "subscriptions": [
                "sub1",
                "sub3",
                "sub4"
             ]
          }
       ],
       "name": "UserB",
       "token": "S3CR3T2",
       "email": "foo-email",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z",
       "created_by": "UserA"
    },
    {
       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bberr",
       "projects": [
          {
             "project": "ARGO",
             "roles": [
                "publisher",
                "consumer"
             ],
             "topics": [
                "topic3"
             ],
             "subscriptions": [
                "sub2"
             ]
          }
       ],
       "name": "UserX",
       "token": "S3CR3T3",
       "email": "foo-email",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z",
       "created_by": "UserA"
    },
    {
       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbfrt",
       "projects": [
          {
             "project": "ARGO",
             "roles": [
                "publisher",
                "consumer"
             ],
             "topics": [
                "topic2"
             ],
             "subscriptions": [
                "sub3",
                "sub4"
             ]
          }
       ],
       "name": "UserZ",
       "token": "S3CR3T4",
       "email": "foo-email",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z",
       "created_by": "UserA"
    }
 ],
 "nextPageToken": "",
 "totalSize": 5
}
```

### Paginated Request that returns the 2 most recent users

### Example request
```
curl -X GET -H "Content-Type: application/json"
  "https://{URL}/v1/users?key=S3CR3T&pageSize=2"
```

### Responses  
If successful, the response contains a list of the 2 most recently added users

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
    },
    {
       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",
       "projects": [
          {
             "project": "ARGO",
             "roles": [
                "consumer",
                "publisher"
             ],
             "topics": [
                "topic1",
                "topic2"
             ],
             "subscriptions": [
                "sub1",
                "sub2",
                "sub3"
             ]
          }
       ],
       "name": "UserA",
       "token": "S3CR3T1",
       "email": "foo-email",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z"
    }
 ],
 "nextPageToken": "some_token2",
 "totalSize": 5
}
```

### Paginated Request that returns the next 3 users

### Example request
```
curl -X GET -H "Content-Type: application/json"
  "https://{URL}/v1/users?key=S3CR3T&pageSize=3&pageToken=some_token2"
```

### Responses  
If successful, the response contains a list of the next 3 users

Success Response
`200 OK`

```json
{
 "users": [
    {
       "uuid": "94bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",
       "projects": [
          {
             "project": "ARGO",
             "roles": [
                "consumer",
                "publisher"
             ],
             "topics": [
                "topic1",
                "topic2"
             ],
             "subscriptions": [
                "sub1",
                "sub3",
                "sub4"
             ]
          }
       ],
       "name": "UserB",
       "token": "S3CR3T2",
       "email": "foo-email",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z",
       "created_by": "UserA"
    },
    {
       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bberr",
       "projects": [
          {
             "project": "ARGO",
             "roles": [
                "publisher",
                "consumer"
             ],
             "topics": [
                "topic3"
             ],
             "subscriptions": [
                "sub2"
             ]
          }
       ],
       "name": "UserX",
       "token": "S3CR3T3",
       "email": "foo-email",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z",
       "created_by": "UserA"
    },
    {
       "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbfrt",
       "projects": [
          {
             "project": "ARGO",
             "roles": [
                "publisher",
                "consumer"
             ],
             "topics": [
                "topic2"
             ],
             "subscriptions": [
                "sub3",
                "sub4"
             ]
          }
       ],
       "name": "UserZ",
       "token": "S3CR3T4",
       "email": "foo-email",
       "service_roles": [],
       "created_on": "2009-11-10T23:00:00Z",
       "modified_on": "2009-11-10T23:00:00Z",
       "created_by": "UserA"
    }
 ],
  "nextPageToken": "some_token3",
  "totalSize": 5
}
```

### Paginated Request that returns all users that are members of a specific project

### Example request
```
curl -X GET -H "Content-Type: application/json"
  "https://{URL}/v1/users?key=S3CR3T&project=ARGO2"
```

### Responses  
If successful, the response contains a list of all available users that are members in the project ARGO2

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
   "uuid": "99bfd746-4rte-11e8-9c2d-fa7ae01bbebc",
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "consumer",
            "publisher"
         ],
         "topics": [
            "topic1",
            "topic2"
         ],
         "subscriptions": [
            "sub1",
            "sub2",
            "sub3"
         ]
      }
   ],
   "name": "UserA",
   "first_name": "FirstA",
   "last_name": "LastA",
   "organization": "OrgA",
   "description": "DescA",
   "token": "S3CR3T1",
   "email": "foo-email",
   "service_roles": [],
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z"
}
```




### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Manage Users - List a specific user by token
This request lists information about a specific user using user's token as input

### Request
```
GET "/v1/users:byToken/{token}"
```

### Where
- token: the token of the user

### Example request
```
curl -X GET -H "Content-Type: application/json"
  "https://{URL}/v1/users:byToken/S3CR3T1?key=S3CR3T"
```

### Responses  
If successful, the response contains information about the specific user

Success Response
`200 OK`

```json
{
   "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "consumer",
            "publisher"
         ],
         "topics": [
            "topic1",
            "topic2"
         ],
         "subscriptions": [
            "sub1",
            "sub2",
            "sub3"
         ]
      }
   ],
   "name": "UserA",
   "first_name": "FirstA",
   "last_name": "LastA",
   "organization": "OrgA",
   "description": "DescA",
   "token": "S3CR3T1",
   "email": "foo-email",
   "service_roles": [],
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z"
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Manage Users - List a specific user by authentication key
This request lists information about a specific user 
based on the authentication key provided as a url parameter

### Request
```
GET "/v1/users/profile"
```

### Example request
```
curl -X GET -H "Content-Type: application/json"
  "https://{URL}/v1/users/profile?key=S3CR3T1"
```

### Responses  
If successful, the response contains information about the specific user

Success Response
`200 OK`

```json
{
   "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "consumer",
            "publisher"
         ],
         "topics": [
            "topic1",
            "topic2"
         ],
         "subscriptions": [
            "sub1",
            "sub2",
            "sub3"
         ]
      }
   ],
   "name": "UserA",
   "first_name": "FirstA",
   "last_name": "LastA",
   "organization": "OrgA",
   "description": "DescA",
   "token": "S3CR3T1",
   "email": "foo-email",
   "service_roles": [],
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z"
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors


## [GET] Manage Users - List a specific user by UUID
This request lists information about a specific user using user's UUID as input

### Request
```
GET "/v1/users:byUUID/{uuid}"
```

### Where
- uuid: the uuid of the user

### Example request
```
curl -X GET -H "Content-Type: application/json"
  "https://{URL}/v1/users:byUUID/99bfd746-4ebe-11e8-9c2d-fa7ae01bbebc?key=S3CR3T"
```

### Responses  
If successful, the response contains information about the specific user

Success Response
`200 OK`

```json
{
   "uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebc",
   "projects": [
      {
         "project": "ARGO",
         "roles": [
            "consumer",
            "publisher"
         ],
         "topics": [
            "topic1",
            "topic2"
         ],
         "subscriptions": [
            "sub1",
            "sub2",
            "sub3"
         ]
      }
   ],
   "name": "UserA",
   "first_name": "FirstA",
   "last_name": "LastA",
   "organization": "OrgA",
   "description": "DescA",
   "token": "S3CR3T1",
   "email": "foo-email",
   "service_roles": [],
   "created_on": "2009-11-10T23:00:00Z",
   "modified_on": "2009-11-10T23:00:00Z"
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
          "project_admin"
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
 "uuid": "99bfd746-4ebe-11e8-9c2a-fa7ae01bbebc",
 "projects": [
    {
       "project": "ARGO",
       "roles": [
          "project_admin"
       ],
       "topics":[],
       "subscriptions":[]
    }
 ],
 "name": "USERNEW",
 "token": "R4ND0MT0K3N",
 "email": "foo-email",
 "service_roles":[],
 "created_on": "2009-11-10T23:00:00Z",
 "modified_on": "2009-11-10T23:00:00Z",
 "created_by": "UserA"
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
"uuid": "99bfd746-4ebe-11e8-9c2d-fa7ae01bbebz",
 "projects": [
    {
       "project": "ARGO2",
       "roles": [
          "project_admin"
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
"uuid": "99bfd740-4ebe-11e8-9c2d-fa7ae01bbebc",
 "projects": [
    {
       "project": "ARGO2",
       "roles": [
          "project_admin"
       ],
       "topics":[],
       "subscriptions":[]
    }
 ],
 "name": "CHANGED_NAME",
 "token": "R4ND0MT0K3N",
 "email": "foo-email",
 "service_roles":[],
 "created_on": "2009-11-10T23:00:00Z",
 "modified_on": "2009-11-11T10:00:00Z",
 "created_by": "UserA"
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
"uuid": "99bfd746-4ebe-11p0-9c2d-fa7ae01bbebc",
 "projects": [
    {
       "project": "ARGO",
       "roles": [
          "project_admin"
       ],
       "topics":[],
       "subscriptions":[]
    }
 ],
 "name": "USER2",
 "token": "NEWRANDOMTOKEN",
 "email": "foo-email",
 "service_roles":[],
 "created_on": "2009-11-10T23:00:00Z",
 "modified_on": "2009-11-11T12:00:00Z",
 "created_by": "UserA"
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
