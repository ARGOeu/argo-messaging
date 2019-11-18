#Schemas Api Calls

Schemas is a resource that works with topics by validating the published messages.

## [GET] Manage Schemas - Retrieve a Schema
This request retrieves a specific schema under the given project

### Request
```json
GET "/v1/projects/{project_name}/schemas/{schema_name}"
```

### Where
- project_name: Name of the project in which the schema will belong
- schema_name: Name of the schema to be created

### Example request
```json
curl -X GET -H "Content-Type: application/json"
 " https://{URL}/v1/projects/project-1/schemas/schema-1?key=S3CR3T"
```

### Responses  

If successful, the response contains the requested schema.

Success Response
`200 OK`
```json
{
    "uuid": "50811bd1-c94c-4ad7-8f55-a561c6270b50",
    "name": "schema-1",
    "type": "json",
    "schema": {
        "properties": {
            "address": {
                "type": "string"
            },
            "email": {
                "type": "string"
            },
            "name": {
                "type": "string"
            },
            "telephone": {
                "type": "string"
            }
        },
        "required": [
            "name",
            "email"
        ],
        "type": "object"
    }
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [POST] Manage Schemas - Create new Schema
This request creates a new schema

### Request
```json
POST "/v1/projects/{project_name}/schemas/{schema_name}"
```

### Where
- project_name: Name of the project in which the schema will belong
- schema_name: Name of the schema to be created

### Example request
```json
curl -X POST -H "Content-Type: application/json -d POSTDATA"
 " https://{URL}/v1/projects/project-1/schemas/schema-1?key=S3CR3T"
```

### Post body:

```json
{
  "type": "json",
  "schema":{
  		"type": "object",
         "properties": {
          "name":        { "type": "string" },
          "email":        { "type": "string" },
          "address":    { "type": "string" },
          "telephone": { "type": "string" }
         },
        "required": ["name", "email"]
  }
}
```

### Responses  

If successful, the response contains the newly created schema.

Success Response
`200 OK`
```json
{
    "uuid": "50811bd1-c94c-4ad7-8f55-a561c6270b50",
    "name": "schema-1",
    "type": "json",
    "schema": {
        "properties": {
            "address": {
                "type": "string"
            },
            "email": {
                "type": "string"
            },
            "name": {
                "type": "string"
            },
            "telephone": {
                "type": "string"
            }
        },
        "required": [
            "name",
            "email"
        ],
        "type": "object"
    }
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors

## [PUT] Manage Schemas - Update Schema
This request updates the contents of a schema. You can update `one` or `all` of the fields at a time.

### Request
```json
PUT "/v1/projects/{project_name}/schemas/{schema_name}"
```

### Where
- project_name: Name of the project under which the schema belongs
- schema_name: Name of the schema to be updated

### Example request
```json
curl -X PUT -H "Content-Type: application/json -d POSTDATA"
 " https://{URL}/v1/projects/project-1/schemas/schema-1?key=S3CR3T"
```

### Post body:

```json
{
  "type": "json",
  "name": "new-name",
  "schema":{
  		"type": "object",
         "properties": {
          "name":        { "type": "string" },
          "email":        { "type": "string" },
          "address":    { "type": "string" },
          "telephone": { "type": "string" }
         },
        "required": ["name", "email", "address"]
  }
}
```

### Responses  

If successful, the response contains the updated schema.

Success Response
`200 OK`
```json
{
    "uuid": "50811bd1-c94c-4ad7-8f55-a561c6270b50",
    "name": "new-name",
    "type": "json",
    "schema": {
        "properties": {
            "address": {
                "type": "string"
            },
            "email": {
                "type": "string"
            },
            "name": {
                "type": "string"
            },
            "telephone": {
                "type": "string"
            }
        },
        "required": [
            "name",
            "email",
            "address"
        ],
        "type": "object"
    }
}
```

### Errors
Please refer to section [Errors](api_errors.md) to see all possible Errors