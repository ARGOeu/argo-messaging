---
id: api_integrations
title: Integrations/Components
sidebar_position: 10
---

Calls for  components that integrate with argo-web-api and access tenant data.

## Refresh Access Token for a Component Integration

Component administrators can request to refresh access keys for their assigned components within a specific tenant.

```
POST /integrations/components/{COMPONENT}/by-project-name/{PROJECT}/refresh
```

Where `{COMPONENT}` the name of the component account supported, example:
- monbox

Where `{PROJECT}` the name of the project

### Request headers

```
Accept: application/json
x-api-key: component-admin-s3cr3t
```

### Response
Headers: `Status: 200 OK`

### Response Body

Json Response example:
```json
{
    "status": {
        "message": "component api key succesfully renewed",
        "code": "200"
    },
    "data": {
        "api_key": "..."
    }
}
```

## Errors


If a component access account has not yet been set up for a specific tenant, then the response will be:

```json
{
    "status": {
        "message": "Component monbox not configured for project FOO",
        "code": "404"
    }
}

```

### Types of components

Types of component accounts supported:

- engine
- poem-admin
- poem-viewer
- monbox
- probe