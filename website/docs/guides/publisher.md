---
id: publisher
title: Publisher Guide
sidebar_position: 1
---

Publishers can send messages to named-channels called Topics. 

## Before you start

In order to get an account on the ARGO Messaging Service, submit a request through the [ARGO Messaging Service account form](https://ams-register.argo.grnet.gr/)

Upon account approval, you will receive information via e-mail about your new project along with an API token.

## Start publishing

When everything is set up you can start by following the general flow for a publisher:

**Step 1:** Create a topic

For more details visit section [Topics: Create a topic](/api_advanced/api_topics.md#create-topic)

**Step 2:** Create a subscription

A Topic without at least one Subscription act like black holes. Publishers can send messages to those topics, but the messages will not be retrievable. In order to be able to publish and consume messages, at least one Subscription must created to the Topic that you are publishing messages to. By default, a Subscription is created in pull mode, meaning that consumers can query the Messaging API and retrieve the messages that are published to the Topic that the Subscription is configured for. More information about how create a Subscription, visit section [Subscriptions: Create a subscription](/api_advanced/api_subs.md#create-subs)

**Step 3:** Start publishing messages

The ARGO Messaging Service accepts JSON over HTTP. In order to publish messages you have to represent them using the following schema:

```json
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "properties": {
    "messages": {
      "type": "array",
      "items": {
        "type": "object",
        "anyOf": [{
          "properties": {
            "data": {
              "type": "string",
              "contentEncoding": "base64",
              "minLength": 1
            },
          },
          "required": ["data"]
        },{
          "properties": {
            "attributes": {
              "type": "object",
              "minProperties": 1,
              "properties": {}
            }
          },
          "required": ["attributes"]
        }]
      }
    }
  },
  "required": [
    "messages"
  ]
}
```

The JSON body send to the ARGO Messaging Service may contain one or more messages. Each message can have:

 - attributes: optional key value pair of metadata you desire
 - data: the data of the message.

The data must be base64-encoded, and can not exceed 10MB after encoding. Note that the message payload must not be empty; it must contain either a non-empty data field, or at least one attribute.

Below you can find an example, in which a user publishes two messages in one call:

```json
{
  "messages": [
  {
    "attributes":
    {
      "station":"NW32ZC",
      "status":"PROD"
    },
    "data":"U28geW91IHdlbnQgYWhlYWQgYW5kIGRlY29kZWQgdGhpcywgeW91IGNvdWxkbid0IHJlc2lzdCBlaCA/"
  },
  {
    "attributes":
    {
      "station":"GHJ32",
      "status":"TEST"
    },
    "data":"U28geW91IHdlbnQgYWhlYWQgYW5kIGRlY29kZWQgdGhpcywgeW91IGNvdWxkbid0IHJlc2lzdCBlaCA/"
  }
  ]
}
```

You can publish and consume any kind of data through the ARGO Messaging Service (as long as the base64 encoded payload is not larger than the maximum acceptable size).

For more details visit section [Topics: Publish message/s to a topic](/api_advanced/api_topics.md#publish)

## Schema Support 

The AMS supports “Schema Validation per topic”. 

When a user want to support a predefined format for messages exchanged then a topic should be created with a schema defined for it.
A schema is a format that messages from a topic must follow. It is actually a contract between publisher and subscriber. The Messaging Service allows the user to define a schema for each topic and validate messages as they are published. It can protect topics from garbage, incomplete messages especially when a topic has multiple remote publishers to ensure data integrity on the client side.

The Schema Support is on demand mechanism that enables a) the definition of the expected payload schema, b) the definition of the expected set of attributes and values and c) the validation for each message if the requirements are met and immediately notify client

The steps that you should follow for a schema support 

**Step 1:** Create a new schema in your project

The Supported Schema Types are JSON and AVRO

For more details visit section  [Create new schema](/api_advanced/api_schemas.md#create-schema)

**Step 2:** Create a topic with this schema attached

If you need to link a schema with your topic you need to provide its name, to the api call during the creation of the topic..

For more details visit section [Create new topic](/api_advanced/api_topics.md#create-topic) 
 
**Step 3:** Assign this schema to your topic 

If you need to link a schema with your topic you need to provide its name, to the api call

For more details visit section [Update the topic](/api_advanced/api_topics.md#create-topic) 

**Step 4:** Validate the message 

This  is used whenever we want to test a message against a schema. The process to check that your schema and messages are working as expected is to create a new topic that needs to be associated with the schema, then create the message in base64 encoding and publish it to the topic. Instead of creating all this pipeline in order to check your schema and messages we can explicitly do it on this API call.

For more details visit section [Validate the message](/api_advanced/api_schemas.md#validate)  

**Step 5:** Publish messages to your topic 

You may now start publishing messages to your topic.

For more details visit section [publish-messages-to-a-topic](/api_advanced/api_topics.md#publish)

