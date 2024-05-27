#Registrations API Calls

ARGO Messaging Service supports calls for registering users

## [POST] Manage Registrations - New user registration
This request creates a new registration for a future user

### Request
```
POST "/v1/registrations
```

### Post body:
```json
{
   "name": "new-register-user",
  "first_name": "first-name",
  "last_name": "last-name",
  "email": "test@example.com",
  "organization": "org1",
  "description": "desc1"
}
```



### Example request
```bash
curl -X POST -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
"https://{URL}/v1/registrations
```

### Responses  
If successful, the response contains the newly registered user

Success Response
`200 OK`

```json
{
   "uuid": "99bfd746-4ebe-11p0-9c2d-fa7ae01bbebc",
   "name": "new-register-user",
   "first_name": "first-name",
   "last_name": "last-name",
   "organization": "org1",
   "description": "desc1",
   "email": "test@example.com",
   "activation_token": "a-token",
   "status": "pending",
   "registered_at": "2009-11-10T23:00:00Z",
   "modified_at": "2009-11-10T23:00:00Z",
   "modified_by": "UserA"
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [POST] Manage Registrations - Accept a User's Registration
This request accepts a user's registration 
and as a result it creates a new user with the provided information.

### Request
```
POST "/v1/registrations/{uuid}:accept"
```

### Example request
```bash
curl -X POST -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
"https://{URL}/v1/registrations/uuid1:accept"
```

### Responses  
If successful, the response contains the newly created user

Success Response
`200 OK`

```json
{
    "uuid": "1d0aa54e-44b8-4d2a-8cf7-d4cb2e350c61",
    "projects": [],
    "name": "user-acc-344",
    "first_name": "fname",
    "last_name": "lname",
    "organization": "grnet",
    "description": "simple user",
    "token": "bb0ad3da48f69372e38e55e423324b7366e32804",
    "email": "test@example.com",
    "service_roles": [],
    "created_on": "2020-05-17T22:27:09Z",
    "modified_on": "2020-05-17T22:27:09Z"
}
```
### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [POST] Manage Registrations - Decline a User's Registration
This request declines a user's registration.
You can also provide a comment regarding
the decline reason of the registration.

### Request
```
POST "/v1/registrations/{uuid}:decline"
```
### Post body:
```json
{
   "comment": "comment"
}
```


### Example request
```bash
curl -X POST -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
"https://{URL}/v1/registrations/uuid1:decline"
```

### Responses  
If successful, the response contains nothing

Success Response
`200 OK`

```json
{}
```
### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Manage Registrations - Retrieve a User's Registration
This request retrieves a user's registration 

### Request
```
GET "/v1/registrations/{uuid}"
```

### Example request
```bash
curl -X GET -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
"https://{URL}/v1/registrations/ur-uuid1"
```

### Responses  
If successful, the response contains user's registration

Success Response
`200 OK`

```json
{
   "uuid": "ur-uuid1",
   "name": "urname",
   "first_name": "urfname",
   "last_name": "urlname",
   "organization": "urorg",
   "description": "urdesc",
   "email": "uremail",
   "status": "pending",
   "activation_token": "uratkn-1",
   "registered_at": "2019-05-12T22:26:58Z",
   "modified_by": "UserA",
   "modified_at": "2020-05-15T22:26:58Z"
}
```
### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [GET] Manage Registrations - Retrieve all registrations
This request retrieves all registration in the service

### Request
```
GET "/v1/registrations"
```

### Optional Filters

- status
- activation_token
- email
- organization
- name

### Example request
```bash
curl -X GET -H "Content-Type: application/json"  -H "x-api-token:S3CR3T" 
"https://{URL}/v1/registrations"
```

### Responses  
If successful, the response contains all registrations

Success Response
`200 OK`

```json
{
   "user_registrations": [
   {
      "uuid": "ur-uuid1",
      "name": "urname",
      "first_name": "urfname",
      "last_name": "urlname",
      "organization": "urorg",
      "description": "urdesc",
      "email": "uremail",
      "status": "pending",
      "activation_token": "uratkn-1",
      "registered_at": "2019-05-12T22:26:58Z",
      "modified_by": "UserA",
      "modified_at": "2020-05-15T22:26:58Z"
   }
  ]
}
```
### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
