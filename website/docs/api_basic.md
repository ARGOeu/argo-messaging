---
id: api_basic
title: Service introduction and configuration
---


## Introduction
The ARGO Messaging Service API implements the Google PubSub specification and thus supports HTTP RPC-style methods in the form of:

 `https://messaging.argo.grnet.gr/api/projects/METHOD`

All methods must be called using HTTPS. Arguments can be passed as GET or POST params, or a mix. The response contains a `200 OK` for a successful request and a JSON object in case of an error. For failure results, the error property will contain a short machine-readable error code. In the case of problematic calls,  during handling user’s request the API responds using a predefined schema (described in chapter Errors), that contains a short machine-readable warning code, an error code and an error description  (or list of them, in the case of multiple errors).

Each user is authenticated by adding the url parameter `?key=T0K3N` in each API request

## Configuration file: config.json

The first step for using the messaging API is to edit the main configuration file.

The ARGO Messaging Service main configuration file is config.json. An example configuration is listed below:

```json
{
  "bind_ip":"",
  "port":8080,
  "zookeeper_hosts":["localhost"],
  "kafka_znode":"",
  "store_host":"localhost",
  "store_db":"argo_msg",
  "certificate":"/etc/pki/tls/certs/localhost.crt",
  "certificate_key":"/etc/pki/tls/private/localhost.key",
  "per_resource_auth":true,
  "service_token":"S0M3T0K3N",
  "log_level":"INFO",
  "log_facilities": ["syslog", "console"]
}
```

### Explanation of config parameters:

Parameter | Description
--------- | -----------
bind_ip | the ip address to listen to.
port | The port where the API will listen to
zookeeper_hosts | List of zookeeper instances that are used to sync kafka
kafka_znode | The znode under which Kafka writes its data on Zookeeper. Default is "" meaning the root node
store_host | Address:port of the datastore server
store_db | Database name used on the datastore server
certificate | path to the node's TLS certificate file
certificate_key | path to the certificate's private key
per_resource_auth | enable authorization per resource (topic/subscription)
service_token | (optional) If set, enables full service-wide access to the api to initialize projects,users and resources
log_level | set the desired log level (defaults to "INFO")
log_facilities | logging output, if left empty, it defaults to console)

**Location of config.json**: API will look first for config.json locally in the folder where the executable runs and then in the ` /etc/argo-messaging/`  location.


## Command line parameters
Apart from configuration file, argo-messaging service accepts configuration parameters in the command line. The list of the available command line parameters is displayed
if the user issues
```
./argo-messaging-service --help
```
The available command line parameters are listed as follows:
```
--bind-ip string           ip address to listen to (default "localhost")
--certificate string       certificate file *.crt (default "/etc/pki/tls/certs/localhost.crt")
--certificate-key string   certificate key file *.key (default "/etc/pki/tls/private/localhost.key")
--config-dir string        directory path to an alternative json config file
--kafka-znode string       kafka zookeeper node name
--log-level string         set the desired log level
--per-resource-auth        enable per resource authentication (default true)
--port int                 port number to listen to (default 8080)
--service-key string       service token definition for immediate full api access
--store-db string          datastore (mongodb) database name (default "argo_msg")
--store-host string        datastore (mongodb) host (default "localhost")
--zookeeper-hosts value    list of zookeeper hosts to connect to (default [localhost])
```

User can optionally specifiy an alternative configuration file directory with the use of the `--config-dir` parameter
For example:
```
./argo-messaging-service --config-dir=/root/alternative/config/
```
The `/root/alternative/config/config.json` must exist
