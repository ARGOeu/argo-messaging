---
id: subscriber_guide
title: Subscriber Guide
sidebar_position: 2
---

Subscribers can read messages from named-channels called Subscriptions.  Each subscription can belong to a single topic. A topic though can have multiple subscriptions. 
If you are a subscriber and you want to receive messages published to a topic, the idea is that you should create a subscription to that topic. 
The subscription is the connection of the topic to a specific application, and its function is to receive and process messages published to the topic. 
Only messages published to the topic after the subscription is created are available to subscriber applications. 


## Before you start

In order to get an account on the ARGO Messaging Service, submit a request through the [ARGO Messaging Service account form](https://ams-register.argo.grnet.gr)

Upon account approval, you will receive information via e-mail about your new project along with an API token.


## Pull vs Push Subscriptions
AMS supports both push and pull message delivery. In push delivery, the Messaging Service initiates requests to your subscriber application to deliver messages. In pull delivery, your subscription application initiates requests to the Pub/Sub server to retrieve messages.

### Pull subscriptions
Pull subscriptions can be configured to require that message deliveries are acknowledged by the Subscribers. If an acknowledgement is made, subscription can resume progressing and send the next available messages. If no acknowledgement is made subscription pauses progressing and re-sends the same messages. In a pull subscription, the subscribing application explicitly calls the API pull method, which requests delivery of a message in the subscription queue. The Pub/Sub server responds with the message (or an error if the queue is empty), and an ack ID. The subscriber then explicitly calls the acknowledge method, using the returned ack ID, to acknowledge receipt.

**GUIDE** [Receiving Messages via Pull](subscriber-pull_guide.md)

### Push subscriptions
In a push subscription, the push server sends a request to the subscriber application, at a preconfigured endpoint. The subscriber's HTTP response serves as an implicit acknowledgement: a success response indicates that the message has been successfully processed and the Pub/Sub system can delete it from the subscription; a non-success response indicates that the Pub/Sub server should resend it (implicit "nack"). To ensure that subscribers can handle the message flow, the Pub/Sub dynamically adjusts the flow of requests and uses an algorithm to rate-limit retries. The push server(s) are an optional set of worker-machines that are needed when the AMS wants to support push enabled subscriptions. It allows to decouple the push functionality from AMS api nodes They perform the push functionality for the messages of a push enabled subscription (consume->deliver→ack)/ Provide a gRPC interface in order to communicate with their api Provide subscription runtime status

**GUIDE** [Receiving Messages via Push](subscriber-push_guide.md)


