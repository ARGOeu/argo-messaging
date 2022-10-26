---
id: mattermost-integration_guide
title: Mattermost Integration
sidebar_position: 7
---

## Overview

Push enabled subscriptions provide us with the functionality to
forward messages to mattermost channels via mattermost webhooks.


#### Mattermost Configuration

Refer to this guide on how to set up your mattermost webhook.
[https://mattermost.com/blog/mattermost-integrations-incoming-webhooks/](https://mattermost.com/blog/mattermost-integrations-incoming-webhooks/)

#### Subscription Configuration

```json
{
  "topic": "projects/example/topics/alarms-reformat-mattermost-topic",
  "pushConfig": {
    "type": "mattermost",
    "maxMessages": 1,
    "retryPolicy": {
      "type": "linear",
      "period": 3000
    },
    "mattermostUrl": "https://example.com/hooks/z5xjq7hzn7yobnjhthrh4q6oxw",
    "mattermostUsername": "bot argo",
    "mattermostChannel": "monitoring-alarms",
    "base64Decode": true
  }
}
```

- `mattermostUrl`: Is the webhook url that will be generated through
the integrations tab of the mattermost UI.

- `mattermostUsername`: Is the username that will be displayed alongside
the forwarded messages.

- `mattermostChannel`: Is the channel that the messages will be forwarded to.

- `base64Decode`: Messages in AMS should be base64 encoded.This flag allows a subscription
to know if the the messages should be first decoded before being pushed
to the remote destination.
  Refer to the following guides to better understand push enabled subscriptions
  and how to use them.

[Swagger Create Subscription](http://argoeu.github.io/argo-messaging/openapi/explore#/Subscriptions/put_projects__PROJECT__subscriptions__SUBSCRIPTION_)

[Push Enabled Subscriptions](http://argoeu.github.io/argo-messaging/docs/api_advanced/api_subscriptions#push-enabled-subscriptions)


## Reformat Messages Example

In some cases, a topic that has some raw messages, but we first
want to process them and reformat them, before pushing to mattermost,
or reusing them for any other activity.
In order to achieve this we need to consume from the topic's subscription
and republish them to another topic after the messages have been processes.
We then attach a push enabled subscription to the topic with the
reformatted messages.

The following snipper shows this kind of functionality.

**NOTE:** Implement your own `format_message()` function to
transform messages to the desired format. The function accepts the
original message decoded as input, and returns the formatted string.

```python

    # set up the ams client
    ams_host = "{0}:{1}".format(args.host, str(args.port))
    LOGGER.info("Setting up AMS client for host {0} and project: {1}".format(ams_host, args.project))
    ams = ArgoMessagingService(endpoint=ams_host, project=args.project, token=args.token)

    while True:
        try:
            # consume alerts
            consumed_messages = ams.pull_sub(sub=args.sub, return_immediately=True, verify=args.verify)
            if len(consumed_messages) == 0:
                time.sleep(args.interval)
                continue
            payload = consumed_messages[0][1].get_data()
            ack_id = consumed_messages[0][0]

            # if we can't parse the message body we should ack the message and move to the next
            try:
                payload = json.loads(payload)
                LOGGER.info("Examining new message {0} . . .".format(ack_id))

                # skip messages that don't have a type of 'endpoint' or 'group'
                if "type" not in payload or (payload["type"] != 'endpoint' and payload["type"] != 'group'):
                    LOGGER.info("Skipping message {0} with wrong payload . . .".format(ack_id))
                    try:
                        ams.ack_sub(sub=args.sub, ids=[ack_id], verify=args.verify)
                        continue
                    except AmsException as e:
                        LOGGER.error("Could not skip message {0}.{1}".format(ack_id, str(e)))
                        continue
            except Exception as e:
                LOGGER.error("Cannot parse payload for message {0}.{1}.Skipping . . .".format(ack_id, str(e)))
                try:
                    ams.ack_sub(sub=args.sub, ids=[ack_id], verify=args.verify)
                    continue
                except AmsException as e:
                    LOGGER.error("Could not skip message {0}.{1}".format(ack_id, str(e)))
                    continue

            # format and publish the new message
            formatted_message = format_message(payload)
            try:
                ams.publish(topic=args.topic, msg=[AmsMessage(data=formatted_message)], verify=args.verify)
            except AmsException as e:
                LOGGER.error("Could not publish to topic.{0}".format(str(e)))
                continue

            # ack the original alert
            try:
                ams.ack_sub(sub=args.sub, ids=[ack_id], verify=args.verify)
            except AmsException as e:
                LOGGER.error("Could not ack original alert {0}.{1}".format(ack_id, str(e)))
        except AmsException as e:
            LOGGER.error("Cannot pull from subscription.{0}".format(str(e)))

        time.sleep(args.interval)
```