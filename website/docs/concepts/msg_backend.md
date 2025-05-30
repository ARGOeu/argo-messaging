---
id: msg_backend
title: Using Apache Kafka as a Backend Message system
sidebar_position: 2
---

The ARGO Messaging API has been designed to rely on a generic Message Back-end Interface and use specific implementation
of that interface for supporting different systems. Right now the first implementation for the messaging back-end relies
on Apache Kafka as a distributed messaging system.

A big advantage of the ARGO Messaging API is that provides a mechanism to easily support namespacing and different
tenants on a Kafka Back-end (Apache Kafka doesn’t support natively namespacing yet). ARGO Messaging API uses the notion
of ‘projects’ for each tenant and can support multiple projects each one containing multiple topics/subscriptions and
users on the same Kafka back-end.
