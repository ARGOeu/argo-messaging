---
id: replaying_guide
title: Replaying Messages
sidebar_position: 5
---

Subscriptions’ messages can still be accessed despite the fact that a Subscriber might have acknowledged them.

This functionality is possible through the subscription’s offset modification. Each subscriptions holds three indices(offsets) that describe the messages that is providing (max:300, min:220, current: 288).

Whenever a message is acknowledged the current offset is incremented, indicating to the subscriber that the next message is available for consumption.

In addition AMS provides the subscriber with the ability to seek offsets for a specific timestamp,the API will provide the closest possible offset it can find for the provided timestamp.

Now that the subscriber has managed to retrieve the subscription’s offset, we can use the modifyOffset api call  to move the indices around and re-consume/replay a subscription’s messages.

For example, if we have the offsets(max:300, min:220, current: 288), moving the current offset to 285, will allow the subscriber to again consume the messages [286,287,288].

Last but not least, it is important to note that a message is being kept available through the AMS api for 7 days. After that time period has passed, it is no longer available and no offset can access it.


## Before you start

In order to get an account on the ARGO Messaging Service, submit a request through the [ARGO Messaging Service account form](https://ams-register.argo.grnet.gr/)

Upon account approval, you will receive information via e-mail about your new project along with an API token.

## Get Subscription's offsets

A subscription’s offsets can be accessed through the API using the following http call [Get Subscription's offsets](api_advanced/api_subs.md#get-offsets).

## Get Subscription's offsets by timestamp

The following http call gives access to the aforementioned functionality [Get Subscription's offsets by timestamp](api_advanced/api_subs.md#get-offset-timestamp).

## Move Subscription's offsets

The following http call gives access to the modifyOffset api call [Move Subscription's offsets](api_advanced/api_subs.md#modify-offsets) to move the indices around and re-consume/replay a subscription’s messages.
