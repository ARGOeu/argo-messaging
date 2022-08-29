---
id: what_is_ams
title: AMS - The Service 
slug: /
---

The ARGO Messaging Service (AMS)  is a Publish/Subscribe Service, which implements the Google PubSub protocol. Instead of focusing on a single Messaging API specification for handling the logic of publishing/subscribing to the broker network the API focuses on creating nodes of Publishers and Subscribers as a Service. It provides an HTTP API that enables Users/Systems to implement message oriented service using the Publish/Subscribe Model over plain HTTP.

The ARGO Messaging Service is a real-time messaging service that allows the user to send and receive messages between independent applications. It is implemented as a Publish/Subscribe Service. Instead of focusing on a single Messaging service specification for handling the logic of publishing/subscribing to the broker network the service focuses on creating nodes of Publishers and Subscribers as a Service. In the Publish/Subscribe paradigm, Publishers are users/systems that can send messages to named-channels called Topics. Subscribers are users/systems that create Subscriptions to specific topics and receive messages. 


## Features 
 - **Ease of use**: It supports an HTTP API and a python library so as to easily integrate with the AMS. 
 - **Push Delivery**: ΑΜS instantly pushes asynchronous event notifications when messages are published to the message topic. Subscribers are notified when a message is available.
 - **Replay messages**: replay messages that have been acknowledged by seeking to a timestamp. 
 - **Schema Support**: on demand mechanism that enables a)  the definition of the expected payload schema, b)  the definition of the expected set of attributes and values and c) the validation for each message if the requirements are met and immediately notify client
 - **Replicate messages on multiple topics**: Republisher script that consumes and publishes messages for specific topics (ex. SITES) 

The AMS supports **“Schema Validation per topic”**. It allows the user to define a schema for each topic and validate messages as they are published. It can protect topics from garbage, incomplete messages especially when a topic has multiple remote publishers to ensure data integrity on the client side. 

The “Replay messages” feature is an offset manipulation mechanism that allows the client on demand to replay or skip messages. When creating a subscription (or editing an existing one), there is an internal option to retain acknowledged messages (by default up to 7 days, or more on request).  To replay and reprocess these messages (ex. testing, error in manipulation etc) , the client has the ability to go back and use the same messages just by seeking a previous timestamp. If  the user needs to skip messages,  he just has to  seek an offset in the future. 

The implementation for the “push server” is one of the features used.  The push server(s) are an optional set of worker-machines - deployed on demand - that are needed when the AMS wants to support push enabled subscriptions. The latest implementation provides a gRPC interface in order to communicate with AMS api.  A new security approach was also introduced to enable a secure handshake. 


## Architectural aspect
 - **Durability**: provide very high durability, and at-least-once delivery, by storing copies of the same message on multiple servers.
 - **Scalability**: It can handle increases in load without noticeable degradation of latency or availability
 - **Latency**: A high performance service that can serve more than 1 billion messages per year 
 - **Availability**:  it deals with different types of issues, gracefully failing over in a way that is unnoticeable to end users. Failures can occur in hardware, in software, and due to load.  

## Fundamentals

In the Publish/Subscribe paradigm, Publishers are users/systems that can send messages to named-channels called Topics. Subscribers are users/systems that create Subscriptions to specific topics and receive messages.

 - **Topics**: Topics are resources that can hold messages. Publishers (users/systems) can create topics on demand and name them (Usually with names that make sense and express the class of messages delivered in the topic)
 - **Subscriptions**: In order for a user to be able to consume messages, he must first create a subscription. Subscriptions are resources that can be created by users on demand and are attached to specific topics. Each topic can have multiple subscriptions but each subscription can be attached to just one topic. Subscriptions allows Subscribers to incrementally consume messages, at their own pace, while the progress is automatically tracked for each subscription.
 - **Message**: The combination of data and (optional) attributes that a publisher sends to a topic and is eventually delivered to subscribers.
 - **Message attribute**: A key-value pair that a publisher can define for a message. 

### Pull vs Push Subscriptions
AMS supports both push and pull message delivery. In push delivery, the Messaging Service initiates requests to your subscriber application to deliver messages. In pull delivery, your subscription application initiates requests to the Pub/Sub server to retrieve messages.

#### Pull subscriptions

Pull subscriptions can be configured to require that message deliveries are acknowledged by the Subscribers. If an acknowledgement is made, subscription can resume progressing and send the next available messages. If no acknowledgement is made subscription pauses progressing and re-sends the same messages.
In a pull subscription, the subscribing application explicitly calls the API pull method, which requests delivery of a message in the subscription queue. The Pub/Sub server responds with the message (or an error if the queue is empty), and an ack ID. The subscriber then explicitly calls the acknowledge method, using the returned ack ID, to acknowledge receipt.

#### Push subscriptions**

In a push subscription, the push server sends a request to the subscriber application, at a preconfigured endpoint. The subscriber's HTTP response serves as an implicit acknowledgement: a success response indicates that the message has been successfully processed and the Pub/Sub system can delete it from the subscription; a non-success response indicates that the Pub/Sub server should resend it (implicit "nack"). To ensure that subscribers can handle the message flow, the Pub/Sub dynamically adjusts the flow of requests and uses an algorithm to rate-limit retries.
The push server(s) are an optional set of worker-machines that are needed when the AMS wants to support push enabled subscriptions.
It allows to decouple the push functionality from AMS api nodes
They perform the push functionality for the messages of a push enabled subscription (consume->deliver→ack)/
Provide a gRPC interface in order to communicate with their api
Provide subscription runtime status
 
**Apart from all these the Messaging Service supports:**

 - **Argo-ams-library**: A simple library to interact with the ARGO Messaging Service.
 - **Argo-AuthN**: Argo-authn is a new Authentication Service. This service provides the ability to different services to use alternative authentication mechanisms without having to store additional user info or implement new functionalities. The AUTH service holds various information about a service’s users, hosts, API urls, etc, and leverages them to provide its functionality.
 - **AMS Metrics**: Metrics about the service and the usage.
 
