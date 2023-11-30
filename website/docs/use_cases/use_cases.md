---
id: use_cases
title: Use cases
sidebar_position: 1
---

# Use cases for the Argo Messaging Service

The integration between different core services using the ARGO Messaging Service (AMS) as transport layer was one of our
main goals. The main services are:

1) **_EOSC Marketplace (beta)_**: It uses the AMS Service to exchange information about the
   orders.

2) **_AAI Federation Registry (beta)_**: It uses the AMS Service to exchange information with the different
   deployers (ex, SimpleSamlPhp, Mitre Id, Keycloak).

3) **_Operations Portal_**: Reads the alarms from predefined topics, stores
   them in a database and displays them in the operations portal.

4) **_Accounting_**: Use of AMS as a transport layer for
   collecting accounting data from the Sites. The accounting information is gathered from different collectors into a
   central accounting repository where it is processed to generate statistical summaries that are available through the
   EGI Accounting Portal.

5) **_FedCloud_**: Use of AMS as a transport layer of the cloud information system. It makes use of the
   ams-authN. The entry point for users, topics and subscriptions is GOCDB.

6) **_ARGO Availability and Reliability Monitoring
   Service_**: It uses the AMS service to send the messages from the monitoring engine to other components.

### AAI Federation Registry Integration

The Federation Registry is a portal designed to manage service providers (SPs). It enables service owners to configure
federated access for their services using the OIDC and SAML protocols by providing a centralized location for managing
the service configuration.
Access management is handled by a different component which can differ based on the installation (Keycloak, SSP,
MitreID). Service configurations have to be updated on the Access Managment component every time a change is made using
the Federation Registry Portal.
Argo messaging is the message-oriented middleware technology that is used for this communication between the two
parties.
The use of Argo Messaging Service allows for:

- **_Flexibility_**: It enables interoperability and integration between different components and systems, regardless of
  their
  underlying technologies or platforms.
- **_Asynchronous communication_**: Messages can be sent and received at different times and speeds, without blocking or
  waiting
  for a response. This improves the responsiveness throughout the system.
- **_Security_**: It provides built-in security features, such as authentication, authorization, encryption, and digital
  signatures, which help protect the confidentiality, integrity, and authenticity of the messages exchanged between the
  Federation Registry and the given Component.

Federation Registry has multiple instances and with the integration of the Argo Messaging Service we have managed to
organise and monitor
communication between our components.
Configuring Argo Messaging and managing topics and subscription was made easy through the use of AMS Admin Ui app and
through the API.

### Live Updates through our Mattermost integration

While the Argo Messaging Service is primarily used for scenarios
where data is being published by one entity and consumed by another,
in order for systems to achieve async event based workflows,the existence
of [push enabled subscriptions](../api_advanced/api_subs.md#push-enabled-subscriptions), gives the ability
to the system itself,
to forward messages to remote destination when they arrive, without having clients
constantly asking for new data.

One use case of this flow, is the ability to deliver data to a specific
mattermost channel.

We have an [mattermost example](../guides/mattermost-integration_guide.md) that mirrors a real use case
where we needed to filter and reformat
specific messages that were actual alerts, that also needed to be delivered
to a mattermost channel in order for issues to be handled as fast as possible.
