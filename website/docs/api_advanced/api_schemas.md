---
id: api_schemas
title: Schemas
sidebar_position: 7
---


Schemas is a resource that works with topics by validating the published messages.

## [GET] Manage Schemas - Retrieve a Schema

This request retrieves a specific schema under the given project

### Request

```
GET "/v1/projects/{project_name}/schemas/{schema_name}"
```

### Where

- project_name: Name of the project in which the schema will belong
- schema_name: Name of the schema to be created

### Example request

```
curl -X GET -H "Content-Type: application/json" -H "x-api-key: S3CR3T"
 "https://{URL}/v1/projects/project-1/schemas/schema-1"
```

### Responses

If successful, the response contains the requested schema.

Success Response
`200 OK`

```json
{
  "uuid": "50811bd1-c94c-4ad7-8f55-a561c6270b50",
  "name": "projects/project-1/schemas/schema-1",
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

Please refer to section [Errors](/api_basic/api_errors.md) to see all possible Errors

## [GET] Manage Schemas - Retrieve All Schemas

This request retrieves all schemas under the given project.

### Request

```
GET "/v1/projects/{project_name}/schemas"
```

### Where

- project_name: Name of the project in which the schema will belong

### Example request

```
curl -X GET -H "Content-Type: application/json" -H "x-api-key: S3CR3T"
 "https://{URL}/v1/projects/project-1/schemas"
```

### Responses

If successful, the response contains all the schemas of the given project.

Success Response
`200 OK`

```json
{
  "schemas": [
    {
      "uuid": "50811bd1-c94c-4ad7-8f55-a561c6270b50",
      "name": "projects/project-1/schemas/schema-1",
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
    },
    {
      "uuid": "50811bd1-c94c-4ad7-8f55-a561c6270b55",
      "name": "projects/project-1/schemas/schema-2",
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
  ]
}
```

### Errors

Please refer to section [Errors](/api_basic/api_errors.md) to see all possible Errors

## [POST] Manage Schemas - Create new Schema {#create-schema}

This request creates a new schema

### Supported Schema Types

> JSON, AVRO

### Request

```
POST "/v1/projects/{project_name}/schemas/{schema_name}"
```

### Where

- project_name: Name of the project in which the schema will belong
- schema_name: Name of the schema to be created

### Example request

```bash
curl -X POST -H "Content-Type: application/json -d $POSTDATA"
 " https://{URL}/v1/projects/project-1/schemas/schema-1"
```

### Post body:

```json
{
  "type": "json",
  "schema": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string"
      },
      "email": {
        "type": "string"
      },
      "address": {
        "type": "string"
      },
      "telephone": {
        "type": "string"
      }
    },
    "required": [
      "name",
      "email"
    ]
  }
}
```

### Responses

If successful, the response contains the newly created schema.

Success Response
`200 OK`

```
{
    "uuid": "50811bd1-c94c-4ad7-8f55-a561c6270b50",
    "name": "projects/project-1/schemas/schema-1",
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

Please refer to section [Errors](/api_basic/api_errors.md) to see all possible Errors

## [PUT] Manage Schemas - Update Schema

This request updates the contents of a schema. You can update `one` or `all` of the fields at a time.

### Request

```
PUT "/v1/projects/{project_name}/schemas/{schema_name}"
```

### Where

- project_name: Name of the project under which the schema belongs
- schema_name: Name of the schema to be updated

### Example request

```bash
curl -X PUT -H "Content-Type: application/json -d $POSTDATA"
 "https://{URL}/v1/projects/project-1/schemas/schema-1"
```

### Post body:

```json
{
  "type": "json",
  "name": "projects/project-1/schemas/new-name",
  "schema": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string"
      },
      "email": {
        "type": "string"
      },
      "address": {
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
    ]
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
  "name": "projects/project-1/schemas/new-name",
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

Please refer to section [Errors](/api_basic/api_errors.md) to see all possible Errors

## [DELETE] Manage Schemas - Delete Schema

This request deletes a schema.

### Request

```
DELETE "/v1/projects/{project_name}/schemas/{schema_name}"
```

### Where

- project_name: Name of the project under which the schema belongs
- schema_name: Name of the schema to be deleted

### Example request

```bash
curl -X DELETE -H "Content-Type: application/json" -H "x-api-key: S3CR3T"
 "https://{URL}/v1/projects/project-1/schemas/schema-1"
```

### Responses

If successful, the response is empty.

Success Response
`200 OK`

```
```

### Errors

Please refer to section [Errors](/api_basic/api_errors.md) to see all possible Errors

## [POST] Manage Schemas - Validate Message {#validate}

This request is used whenever we want to test a message against a schema.
The process to check that your schema and messages are working as expected is to create
a new topic that needs to be associated with the schema, then create the message in base64 encoding and
publish it to the topic. Instead of creating all this pipeline in order to check your schema and messages
we can explicitly do it on this API call.

### Request

```
POST "/v1/projects/{project_name}/schemas/{schema_name}:validate"
```

### Where

- project_name: Name of the project under which the schema belongs
- schema_name: Name of the schema to be updated

### Example request

```bash
curl -X POST -H "Content-Type: application/json -d $POSTDATA"
 "https://{URL}/v1/projects/project-1/schemas/schema-1:validate"
```

### Post body:

#### JSON Schema

```json
{
  "name": "name1",
  "email": "e1@example.com",
  "address": "address1",
  "telephone": "6980574421"
}
```

#### AVRO Schema

When dealing with an AVRO Schema, the binary message needs to be encoded to `base64`
alongside its `schema` and sent via the `data` field which is required.

```json
{
  "data": "DGFnZWxvc8T8Cg=="
}
```

### Responses

Success Response
`200 OK`

```json
{
  "message": "Message validated successfully"
}
```

### Errors

Please refer to section [Errors](/api_basic/api_errors.md) to see all possible Errors
