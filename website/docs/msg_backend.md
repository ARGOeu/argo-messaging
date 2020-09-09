---
id: msg_backend
title: Using Apache Kafka as a Backend Message system
---

The ARGO Messaging API has been designed to rely on a generic Message Backend Interface and use specific implementation of that interface for supporting different systems. Right now the first implementation for the messaging backend relies on Apache Kafka as a distributed messaging system.

A big advantage of the ARGO Messaging API is that provides a mechanism to easily support namespacing and different tenants on a Kafka Backend (Apache Kafka doesn’t support natively namespacing yet). ARGO Messaging API uses the notion of ‘projects’ for each tenant and can support multiple projects each one containing multiple topics/subscriptions and users on the same Kafka backend.
