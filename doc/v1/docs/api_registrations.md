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
   "registered_at": "2009-11-10T23:00:00Z"
}
```
