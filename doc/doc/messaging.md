---
title: Messaging Service | ARGO
layout: api
page_title: Messaging Service
font_title: 'fa fa-comments-o'
description: This document describes the Messaging service v1.
---

## Description

The ARGO Messaging Service is a Publish/Subscribe Service, which implements the Google PubSub protocol. It provides an HTTP API that enables Users/Systems to implement message oriented service using the Publish/Subscribe Model over plain HTTP.

In the Publish/Subscribe paradigm, Publishers are users/systems that can send messages to named-channels called Topics. Subscribers are users/systems that create Subscriptions to specific topics and receive messages.


## Topics

Topics are resources that can hold messages. Publishers (users/systems) can create topics on demand and name them (Usually with names that make sense and express the class of messages delivered in the topic)

## Subscriptions

In order for a user to be able to consume messages, he must first create a subscription. Subscriptions are resources that can be created by users on demand and are attached to specific topics. Each topic can have multiple subscriptions but each subscription can be attached to just one topic. Subscriptions allows Subscribers to incrementally consume messages, at their own pace, while the progress is automatically tracked for each subscription.

## Pull vs Push Subscriptions

Pub/Sub supports both push and pull message delivery. In push delivery, the Pub/Sub initiates requests to your subscriber application to deliver messages. In pull delivery, your subscription application initiates requests to the Pub/Sub server to retrieve messages.

### Pull subscriptions

Pull subscriptions can be configured to require that message deliveries are acknowledged by the Subscribers. If an acknowledgement is made, subscription can resume progressing and send the next available messages. If no acknowledgement is made subscription pauses progressing and re-sends the same messages.

In a pull subscription, the subscribing application explicitly calls the API pull method, which requests delivery of a message in the subscription queue. The Pub/Sub server responds with the message (or an error if the queue is empty), and an ack ID. The subscriber then explicitly calls the acknowledge method, using the returned ack ID, to acknowledge receipt.

### Push subscriptions

In a push subscription, the Pub/Sub server sends a request to the subscriber application, at a preconfigured endpoint. The subscriber's HTTP response serves as an implicit acknowledgement: a success response indicates that the message has been successfully processed and the Pub/Sub system can delete it from the subscription; a non-success response indicates that the Pub/Sub server should resend it (implicit "nack"). To ensure that subscribers can handle the message flow, the Pub/Sub dynamically adjusts the flow of requests and uses an algorithm to rate-limit retries.

In the current implementation of the AMS there is support only for pull subscriptions. Support for pull subscription will be available in the next versions.

## Messages

In the ARGO Messaging Service each message has an identifier, data (payload) and metadata (optional). The metadata are stored in a attribute dictionary as key/value pairs. 

### Message acknowledgement deadline

The ack deadline is the number of seconds after delivery, during which the subscriber must acknowledge the receipt of a pull or push message. If a subscriber does not respond with an explicit acknowledge (for a pull subscriber) or with a success response code (for a push subscriber) by this deadline, the server will attempt to resend the message. By default this deadline is 10 seconds.


## Security and privacy considerations

Authentication is the process of determining the identity of a client, which is typically a user account. Authorization is the process of determining what permissions an authenticated identity has on a set of specified resources. In the Messaging API, there can be no authorization without authentication.

This is an initial implementation of the user authentication and authorization. In the next versions of the ARGO Messaging service we are going to provide support for both bear and OpenID Connect tokens for the API access and it will be possible to apply ACLs at each (resource) subscriptions and topics.

## Using Apache Kafka as a Backend Message system

The ARGO Messaging API has been designed to rely on a generic Message Backend Interface and use specific implementation of that interface for supporting different systems. Right now the first implementation for the messaging backend relies on Apache Kafka as a distributed messaging system.

A big advantage of the ARGO Messaging API is that provides a mechanism to easily support namespacing and different tenants on a Kafka Backend (Apache Kafka doesn’t support natively namespacing yet). ARGO Messaging API uses the notion of ‘projects’ for each tenant and can support multiple projects each one containing multiple topics/subscriptions and users on the same Kafka backend.

## Links and further reading

- [Swagger : Demo](https://api-doc.argo.grnet.gr/argo-messaging/)
- [Mkdocs : Documentation](/messaging/v1/)

