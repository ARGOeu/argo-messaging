# Publisher Guide

Publishers are users/systems that can send messages to named-channels called Topics. 

## Before you start

The first thing you need to do before starting publishing is to make a request:

 - for a project
 - and for an account (token) to the project

by sending an email to argo-dev@lists.grnet.gr

## Start publishing

When everything is set up you can start by following the general flow for a publisher:

**Step 1: ** Define a topic and send a request to  to create it.

For more details visit section [Topics: Create a topic](api_topics.md#put-manage-topics-create-new-topic)


**Step 2: ** Prepare a message to send.

In order to send messages you must prepare a json of the following format.

```json
{
"messages": [
 	{
  		"attributes": {
        "attr1":"test1",
        "attr2":"test2"
   		}
  	,
 "data":"U28geW91IHdlbnQgYWhlYWQgYW5kIGRlY29kZWQgdGhpcywgeW91IGNvdWxkbid0IHJlc2lzdCBlaCA/"

 	}
]
}
```

Json may contain a message or a list of messages with:

 - attributes: optional key value pair of metadata you desire
 - data: the data of the message. 

The message data must be base64-encoded, and can be a maximum of 10MB after encoding. Note that the message payload must not be empty; it must contain either a non-empty data field, 
or at least one attribute.

For more details visit section [Topics: Publish message/s to a topic](api_topics.md#post-publish-messages-to-a-topic)

**Step 3: ** Send a request to publish the message.

The topic:publish endpoint publishes a message, or a list of messages to a specific topic with a POST request

For more details visit section [Topics: Publish message/s to a topic](api_topics.md#post-publish-messages-to-a-topic)


