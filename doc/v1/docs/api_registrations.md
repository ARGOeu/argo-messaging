#Registrations Api Calls

ARGO Messaging Service supports calls for registering users

## [POST] Manage Registrations - New user registration
This request is

### Request
```json
POST "/v1/registrations/{user_name}"
```

### Post body:
```json
{
  "first_name": "first-name",
  "last_name": "last-name",
  "email": "test@example.com",
  "organization": "org1",
  "description": "desc1"
}
```



### Example request
```bash
curl -X POST -H "Content-Type: application/json"
"https://{URL}/v1/registrations/new-register-name"
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
```json
POST "/v1/registrations/{activation_token}:accept"
```

### Example request
```bash
curl -X POST -H "Content-Type: application/json"
"https://{URL}/v1/registrations/token:accept"
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