# List API Version Information

This method can be used to retrieve api version information

The `release` field is only available if proper authentication credentials are provided.
## Input

```
GET /v1/version
```

### Request headers

```
Accept: application/json
```

## Response

Headers: `Status: 200 OK`

## Response Body

Json Response

```json
{
    "build_time": "2019-11-01T12:51:04Z",
    "golang": "go1.15.6",
    "compiler": "gc",
    "os": "linux",
    "architecture": "amd64",
    "release": "1.1.0",
    "distro": "Ubuntu 20.04 LTS",
    "hostname": "lb.ams.gr"
}
```
