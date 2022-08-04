#!/usr/bin/env python

from argo_ams_library import ArgoMessagingService, AmsException, AmsMessage
import logging.handlers
import sys
import json
import time
import argparse

# set up logging
LOGGER = logging.getLogger("AMS Format Alerts Script")

# emojis
EMOJIS = {
    "CRITICAL": ":red_circle:",
    "WARNING": ":large_orange_circle:",
    "OK": ":large_green_circle:",
    "UNKNOWN": ":white_circle:"
}

# message template
TEMPLATE = "{0} **{1}** is {2} {3} \n**Summary**: {4} \n**URL** : {5}"


def format_message(payload):

    emoji = EMOJIS[payload["status"]]

    return TEMPLATE.format(emoji,payload["endpoint_group"], payload["status"],
                           emoji, payload["summary"], payload["url.history"])


def main(args):

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


if __name__ == "__main__":

    parser = argparse.ArgumentParser(description="Format alerts to be used in mattermost push enabled subscriptions")

    parser.add_argument(
        "-host", "--host", metavar="STRING", help="Ams endpoint", type=str, dest="host", required=True)

    parser.add_argument(
        "-port", "--port", metavar="INTEGER", help="Ams port", default=443, type=str, dest="port")

    parser.add_argument(
        "-token", "--token", metavar="STRING", help="Ams token", type=str, dest="token", required=True)

    parser.add_argument(
        "-project", "--project", metavar="STRING", help="Ams project", type=str, dest="project", required=True)

    parser.add_argument(
        "-sub", "--sub", metavar="STRING", help="Ams sub", type=str, dest="sub", required=True)

    parser.add_argument(
        "-topic", "--topic", metavar="STRING", help="Ams topic", type=str, dest="topic", required=True)

    parser.add_argument(
        "--verify", help="SSL verification for requests", dest="verify", action="store_true")

    parser.add_argument(
        "-interval", "--interval", metavar="INTEGER", type=int, default=0,
        help="How often should it try to publish in seconds", dest="interval")

    console_handler = logging.StreamHandler()
    console_handler.setFormatter(logging.Formatter('%(asctime)s %(name)s[%(process)d]: %(levelname)s - %(message)s'))
    LOGGER.addHandler(console_handler)
    LOGGER.setLevel(logging.INFO)

    sys.exit(main(parser.parse_args()))
