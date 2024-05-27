# Service health check for ams and push server

This method can be used to retrieve api information regarding the proper functionality
of the ams service and the push server

## [GET] Get Health status

### Request
```
GET "/v1/status"
```

### Example request

- `details=(true|false)` indicates if we need detailed
information about errors regarding the push server.

- A user token corresponding to a `service_admin` or `admin_viewer`
has to be provided when using the `details` parameter.

```bash
curl -H "Content-Type: application/json" -H "x-api-token:S3CR3T" 
 "https://{URL}/v1/status?details=true"
```

### Responses
If successful, the response returns the health status of the service

Success Response
`200 OK`

```json
{
  "status": "ok",
  "push_servers": [
    {
      "endpoint": "localhost:5555",
      "status": "Success: SERVING"
    }
  ]
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors
