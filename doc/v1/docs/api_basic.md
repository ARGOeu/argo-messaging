# Introduction
The ARGO Messaging Service API implements the Google PubSub specification and thus supports HTTP RPC-style methods in the form of:

 `https://messaging.argo.grnet.gr/api/projects/METHOD`

All methods must be called using HTTPS. Arguments can be passed as GET or POST params, or a mix. The response contains a `200 OK` for a successful request and a JSON object in case of an error. For failure results, the error property will contain a short machine-readable error code. In the case of problematic calls,  during handling user’s request the API responds using a predefined schema (described in chapter Errors), that contains a short machine-readable warning code, an error code and an error description  (or list of them, in the case of multiple errors).

Each user is authenticated by adding the url parameter `?key=T0K3N` in each API request

## Configuration file: config.json

The first step for using the messaging API is to edit the main configuration file.

The ARGO Messaging Service main configuration file is config.json. An example configuration is listed below:

```json
{
  "port":8080,
  "broker_host":"localhost:9092",
  "store_host":"localhost",
  "store_db":"argo_msg",
  "use_authorization":true,
  "use_authentication":true,
  "use_ack":true
}
```

### Explanation of config parameters:

Parameter | Description
--------- | -----------
port | The port where the API will listen to
broker_host | Address:port of the broker instance
store_host | Address:port of the datastore server
store_db | Database name used on the datastore server
use_authorization | If true, API will boot with support for authorization
use_authentication | If true, API will boot with support for authentication
use_ack | If true, API will boot with acknowledgement support when consuming messages

**Location of config.json**: API will look first for config.json locally in the folder where the executable runs and then in the ` /etc/argo-messaging/`  location.
