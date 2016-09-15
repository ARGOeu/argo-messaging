# Publisher Guide

Publishers can send messages to named-channels called Topics. 

## Before you start

In order to get an account on the ARGO Messaging Service, submit a request through the [ARGO Messaging Service account form](https://docs.google.com/forms/d/e/1FAIpQLScfMCYPkUqUa5lT046RK1yCR4yn6M96WbgD5DMlNJ-zRFHSRA/viewform)

Upon account approval, you will receive information via e-mail about your new project along with an API token.

## Start publishing

When everything is set up you can start by following the general flow for a publisher:

**Step 1:** Create a topic

For more details visit section [Topics: Create a topic](api_topics.md#put-manage-topics-create-new-topic)

**Step 2:** Create a subscription

A Topic without at least one Subscription act like black holes. Publishers can send messages to those topics, but the messages will not be retrievable. In order to be able to publish and consume messages, at least one Subscription must created to the Topic that you are publishing messages to. By default, a Subscription is created in pull mode, meaning that consumers can query the Messaging API and retrieve the messages that are published to the Topic that the Subscription is configured for. More information about how create a Subscription, visit section [Subscriptions: Create a subscription](api_subs.md#put-manage-subscriptions-create-subscriptions)

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

For more details visit section [Topics: Publish message/s to a topic](api_topics.md#post-publish-messages-to-a-topic)

