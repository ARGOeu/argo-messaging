---
id: subscriber-pull_guide
title: Receiving messages using Pull
sidebar_position: 3
---

Subscribers can read messages from named-channels called Subscriptions.  Each subscription can belong to a single topic. A topic though can have multiple subscriptions. 
If you are a subscriber and you want to receive messages published to a topic, the idea is that you should create a subscription to that topic. 
The subscription is the connection of the topic to a specific application, and its function is to receive and process messages published to the topic. 
Only messages published to the topic after the subscription is created are available to subscriber applications. 

In pull delivery, your subscription application initiates requests to the Pub/Sub server to retrieve messages. Pull subscriptions can be configured to require that message deliveries are acknowledged by the Subscribers. If an acknowledgement is made, subscription can resume progressing and send the next available messages. If no acknowledgement is made subscription pauses progressing and re-sends the same messages. In a pull subscription, the subscribing application explicitly calls the API pull method, which requests delivery of a message in the subscription queue. The Pub/Sub server responds with the message (or an error if the queue is empty), and an ack ID. The subscriber then explicitly calls the acknowledge method, using the returned ack ID, to acknowledge receipt.

## Before you start

In order to get an account on the ARGO Messaging Service, submit a request through the [ARGO Messaging Service registration form](https://ams-register.argo.grnet.gr/)

Upon account approval, you will receive information via e-mail about your new project along with an API token.

## Schema Support 

As already mentioned, the AMS supports “Schema Validation per topic”. The subscription consumes the messages from the topic with the defined schema. 

For more information visit [Schemas](api_advanced/api_schemas.md)

## Consume Messages 

AMS Service supports a request that consumes messages from a subscription in a project. It's important to note that the subscription's topic must exist in order for the user to pull messages. At the same time Only messages published to the topic after the subscription is created are available to subscriber applications.
In AMS the request supports the following parameters 

 - maxMessages: the max number of messages to consume
 - returnImmediately: (true or false) to prevent the subscriber from waiting if the queue is currently empty. If not specified the default value is true.

The user may specify the max number of messages returned by one call by setting maxMessages field. By default, the server will keep the connection open until at least one message is received; you can optionally set the returnImmediately field to true to prevent the subscriber from waiting if the queue is currently empty.

For more information visit [Consume messages](api_advanced/api_subs.md#post-pull-messages-from-a-subscription-consume)

## Acknowledge messages

The messages are stored in the queue. In order to remove the messages from the queue they should be Acknowledged. Messages retrieved from a pull subscription can be acknowledged by sending message with an array of ackIDs.

For more information visit [Sending an ack](api_advanced/api_subs.md#post-sending-an-ack)

