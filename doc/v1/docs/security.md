# Security in Argo Messaging service

## Authentication

Authentication in the AMS takes place using an `url key` provided
with each API request.

The large majority of api calls support the `url parameter`, <b>key</b>.

E.g. `/v1/projects?key=b328c7890f061f87cbd4rff34f36fa2ae20993a5`

<b> The service also supports the use of the x-api-key header
for the user to provide its key. </b>

- Each request will extract the key from the request parameters
and will try to find a user associated with it in the respective
data store.

- The key can also be refreshed when needed with the
`/users/{user}:refreshToken` api call.

- API keys are expected to be used by external service's clients.

### X509 Authentication
Although AMS doesn't support direct authentication through an x509 certificate,
you can use the [argo-authentication-service](https://github.com/ARGOeu/argo-api-authn)
to map an x509 certificate to an AMS `key`.
The service will also validate the certificate.
The [ams-library](https://github.com/ARGOeu/argo-ams-library) will effortlessly
hide this complexity if you decide to use it in order to access AMS.


## Authorization

After the authentication part takes place, the user will be also assigned
its privileges/roles in order for the service to determine,
if the `user is allowed to access`
 - the requested resource.
 - perform a certain action.
 
The Argo Messaging Service supports the following `core` roles:

- `service_admin` - which is targeted to users that have an administrative duty over the service.
Service admin is a service wide role.
- `project_admin` -  which is targeted towards users that manage the resources/actions
under a specific ams project.Project admins can only access the project(s) they belong to.
Project admin is a `per project role` not a service wide role.
- `publisher` - which is targeted towards users that primarily publish messages to topics.
Publishers are able to access topic(s) under the project(s) they belong to.
Publisher is a `per project role` not a service wide role.
- `consuner` - which is targeted towards users that primarily consume messages from subscriptions. 
 Consumers are able to access subscriptions(s) under the project(s) they belong to.
 Consumer is a `per project role` not a service wide role.
 
 E.g. `userA` can be a 
 - `project_admin` under `projectA`,
 - `publihser` under `projectB`
 - `publisher` & `consumer` under `projectC`. 
 
 
 Each API route gets assigned which roles it should accept,
 
-  /v1/projects is only accessible by `service_admin`,
 
 -  /v1/topics/{topic}:publish is accessible by `service_admin`, `project_admin` and `publisher`.
 
 ## ACL Based access
 
All publishers `cannot access` all topics under their project.
Same for consumers, they `cannot access` all subscriptions under their project.

Both Topics and Subscriptions have ACLs which determine which of the project's 
`publishers` and `consumers` respectively, can access them.
ACLs for topics and subscriptions contain `user names`.

## Push Enabled Subscriptions

### Verifying Ownership of Push Subscriptions Endpoints
Whenever a subscription is created with a valid push configuration, the service will also generate a unique hash that
should be later used to validate the ownership of the registered push endpoint, and will mark the subscription as 
unverified.This procedure is mandatory in order to avoid spam requests to
endpoints that don't belong to the right user.

The owner of the push endpoint needs to execute the following steps in order to verify the ownership of the
registered endpoint.

- Expose an api call with a path of `/ams_verification_hash`. The service will try to access this path using the `host:port`
of the push endpoint. For example, if the push endpoint is `https://example.com:8443/receive_here`, the  push endpoint should also
support the api route of `https://example.com:8443/ams_verification_hash`.

- The api route of `https://example.com:8443/ams_verification_hash` should support the http `GET` method.

- A `GET` request to `https://example.com:8443/ams_verification_hash` should return a response body 
with only the `verification_hash`
that is found inside the subscriptions push configuration, 
a `status code` of `200` and the header `Content-type: plain/text`.

### Securing remote push endpoints

If you want to secure your remote endpoint, you can have the service generate
a unique authorization hash for the subscription, which means that all
push messages will contain the generated token inside
the `Authorization` header.
As a result the remote endpoint can authenticate incoming
push messages.

## AMS - Push Server Connectivity

AMS doesn't handle the actual pushing of messages for push enabled subscriptions
,only the configuration part of them.

The [ams-push-server](https://github.com/ARGOeu/ams-push-server) 
component is responsible for delivering push messages
to remote endpoints.

AMS and Push server communicate with each other using `mutual TLS`
for authentication, while the push server also implements an authorization strategy
of accepting requests only from certificates that have specific
`Common Name(s)`.