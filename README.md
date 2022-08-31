# ARGO Messaging

The ARGO Messaging Service is a Publish/Subscribe Service,
which implements the Google PubSub protocol. 
Instead of focusing on a single Messaging API specification 
for handling the logic of publishing/subscribing 
to the broker network the API focuses 
on creating nodes of Publishers and Subscribers as a Service.
It provides an HTTP API that enables Users/Systems to implement
message oriented services using the Publish/Subscribe Model over plain HTTP.
In the Publish/Subscribe paradigm, Publishers are users/systems 
that can send messages to
named-channels called Topics. Subscribers are users/systems that
create Subscriptions to
specific topics and receive messages.

## Prerequisites 

#### Build Requirements

 - Golang 1.15

#### Datastore Requirements
  - The service has been tested with mongodb from version `3.2.22` up to `4.2.3`.
 
#### Broker requirements

  - Kafka 2.2.1
  - Zookeeper 3.4.5
  
#### Push Server
In order to support push enabled subscriptions AMS relies on an external service
that handles the actual pushing of messages, while AMS holds the configuration
for the subscriptions.You can create push enabled subscriptions even
when the push-server isn't available, they will be picked up automatically
when the push-server is up and running.
- [Push server](https://github.com/ARGOeu/ams-push-server)


## Configuration

#### Configuration Location
Configuration for the service takes place inside a `config.json` file, that
resides in two possible locations:

1) Same folder as the binary

2) `/etc/argo-messaging/config.json`

#### Configuration values

- `port` - port the service will bind to
- `zookeeper_hosts` - list of zookeeper hosts, e.g. [zoo1:2181,zoo2:2181,zoo3:2181]
- `store_host` - store host, e.g. 'mongo1:27017,mongo2:27017,mongo3:27017'
- `store_db` - mongo db database name
- `certificate` - /path/to/tls/certificate
- `certificate_key` - /path/to/cert/ley
- `certificate_authorities_dir` - dir containing CAs
- `log_level` - DEBUG,INFO,WARNING, ERROR or FATAL
-  `push_enabled` - (true|false) whether or not the service will support push enabled subscriptions
- `push_tls_enabled` - (true|false), whether or not the service will communicate over TLS with the push server
- `push_server_host` - push1.grnet.gr
- `push_server_port` - 443
- `verify_push_server` - (true|false) mutual TLS for the push server
- `push_worker_token` - token for the active push worker user
- `log_facilities` - ["syslog", "console"]  
- `auth_option` - (`key`|`header`|`both`), where should the service look for the access token.
- `proxy_hostname` - The FQDN of any proxy or load balancer that might serve request in place of the AMS

#### Build & Run the service

In order to build the service, inside the AMS repo issue the command:
```bash
go build
```
In order to run the service,
```bash
./argo-messaging
```

## X509 Authentication
Although AMS doesn't support direct authentication through an x509 certificate,
you can use the [argo-authentication-service](https://github.com/ARGOeu/argo-api-authn)
to map an x509 certificate to an AMS `key`.
The service will also validate the certificate.
The [ams-library](https://github.com/ARGOeu/argo-ams-library) will effortlessly
hide this complexity if you decide to use it in order to access AMS.

## Managing the protocol buffers and gRPC definitions

In order to modify any `.proto` file you will need the following

 - Read on how to install the protoc compiler on your platform [here.](https://github.com/protocolbuffers/protobuf)

 -  Install the go plugin. `go get -u github.com/golang/protobuf/protoc-gen-go`

 - install the go gRPC package. `go get -u google.golang.org/grpc`

 - Inside `push/grpc` compile. `protoc -I proto/ proto/ams.proto --go_out=plugins=grpc:proto`

## Helpful utilities

Inside the [tools](https://github.com/ARGOeu/argo-messaging/tree/master/tools) folder you can find various scripts that can help you
perform common tasks OR help you get started with interacting with AMS.

There is also a handy python [library]((https://github.com/ARGOeu/argo-ams-library))
for interacting with AMS.


## Credits

The ARGO Messaging Service is developed by [GRNET](http://www.grnet.gr)

The work represented by this software was partially funded by 
 - EGI-ENGAGE project through the European Union (EU) Horizon 2020 program under Grant number 654142.
 - EOSC-Hub project through the European Union (EU) Horizon 2020 program under Grant number 77753642.
