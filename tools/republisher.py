#!/usr/bin/env python

from avro.io import BinaryEncoder, BinaryDecoder
from avro.io import DatumWriter, DatumReader
import avro.schema
from io import BytesIO
import argo_ams_library
from argo_ams_library import ArgoMessagingService
import argparse
import base64
import logging
import logging.handlers
import sys
import json
import time

# set up logging
LOGGER = logging.getLogger("AMS republish script")


def extract_messages(ams, ingest_sub, bulk_size, schema, verify):

    # consume metric data messages
    consumed_msgs = ams.pull_sub(ingest_sub, num=bulk_size, return_immediately=True, verify=verify)

    # initialise the avro reader
    avro_reader = DatumReader(writers_schema=schema)

    # all the decoded messages that will be returned
    decoded_msgs = []

    # decode the messages
    for msg in consumed_msgs:

        try:

            # decode the data field again using the provided avro schema
            msg_bytes = BytesIO(msg[1].get_data())
            msg_decoder = BinaryDecoder(msg_bytes)
            avro_msg = avro_reader.read(msg_decoder)

            # check that the tags field is present
            if avro_msg["tags"] is None:
                raise KeyError("tags field is empty")

            # append to decoded messages
            decoded_msgs.append((msg[0], avro_msg))

        except Exception as e:
            LOGGER.warning("Could not extract data from ams message {}, {}".format(msg[0], e.message))

    last_msg_id = "-1"
    if len(consumed_msgs) > 0:
        last_msg_id = consumed_msgs.pop()[0]

    return decoded_msgs, last_msg_id


def filter_messages(consumed_msgs, sites):

    filtered_msgs = []

    for msg in consumed_msgs:

        if "endpoint_group" not in msg[1]["tags"]:
            LOGGER.warning("Message {} has no endpoint_group".format(msg[0]))
            continue

        if msg[1]["tags"]["endpoint_group"] in sites:
            filtered_msgs.append(msg)

    return filtered_msgs


def republish_messages(filtered_msgs, ams, verify):

    for msg in filtered_msgs:

        topic = msg[1]["tags"]["endpoint_group"]

        fields = ["status", "service", "timestamp", "metric", "hostname", "monitoring_host"]

        header = dict()
        for fl in fields:
            if msg[1][fl] is None:
                LOGGER.warning("Message {} contains empty field {}".format(msg[0], fl))
                header[fl] = ""
            else:
                header[fl] = msg[1][fl]

        data = dict()
        if msg[1]["summary"] is None:
            LOGGER.warning("Message {} contains no summary field".format(msg[0]))
            data["body"] = ""
        else:
            data["body"] = msg[1]["summary"]

        data["header"] = header
        data["text"] = "true"

        ams_msg = argo_ams_library.AmsMessage(data=json.dumps(data))

        ams.publish(topic, ams_msg, verify=verify)


def main(args):

    # set up the configuration object
    config = dict()

    # default values
    config["bulk_size"] = 100
    config["interval"] = 10

    with open(args.ConfigPath, 'r') as f:
        config = json.load(f)

    # stream(console) handler
    console_handler = logging.StreamHandler()
    console_handler.setFormatter(logging.Formatter('%(asctime)s %(name)s[%(process)d]: %(levelname)s %(message)s'))
    LOGGER.addHandler(console_handler)
    if args.debug:
        LOGGER.setLevel(logging.DEBUG)
    else:
        LOGGER.setLevel(logging.INFO)

    # sys log handler
    syslog_handler = logging.handlers.SysLogHandler(config["syslog_socket"])
    syslog_handler.setFormatter(logging.Formatter('%(asctime)s %(name)s[%(process)d]: %(levelname)s %(message)s'))
    if args.debug:
        syslog_handler.setLevel(logging.DEBUG)
    else:
        syslog_handler.setLevel(logging.INFO)

    syslog_handler.setLevel(logging.INFO)
    LOGGER.addHandler(syslog_handler)

    # start the process of republishing messages

    ams_endpoint = "{}:{}".format(config["ams_host"], config["ams_port"])

    ams = ArgoMessagingService(endpoint=ams_endpoint, token=config["ams_token"], project=config["ams_project"])

    schema = avro.schema.parse(open(config["avro_schema"], "rb").read())

    while True:

        start_time = time.time()

        try:
            consumed_msgs, last_msg_id = extract_messages(ams, config["ingest_subscription"], config["bulk_size"], schema, args.verify)
            if last_msg_id == "-1":
                LOGGER.info("No new messages")
                time.sleep(config["interval"])
                continue

            LOGGER.debug("Consumed messages \n {}".format(consumed_msgs))

            filtered_msgs = filter_messages(consumed_msgs, config["sites"])

            LOGGER.debug("Filtered messages \n {}".format(filtered_msgs))

            republish_messages(filtered_msgs, ams, args.verify)

            # make sure that the acknowledgment happens
            try:
                # try to acknowledge
                ams.ack_sub(config["ingest_subscription"], [last_msg_id], verify=args.verify)
            except Exception as e:
                # if the acknowledgment fails
                LOGGER.critical("Retrying to acknowledge message {} after error {}".format(last_msg_id, e.message))
                while True:
                    try:
                        # consume again in order to refresh the TTL
                        ams.pull_sub(config["ingest_subscription"], config["bulk_size"], True, verify=args.verify)
                        # try to ack again using the msg_id from the first consumption
                        ams.ack_sub(config["ingest_subscription"], [last_msg_id], verify=args.verify)
                        break
                    except Exception as e:
                        LOGGER.critical(
                            "Retrying to acknowledge message {} after error {}".format(last_msg_id, e.message))

                    time.sleep(config["interval"])

            end_time = time.time()

            LOGGER.info("Consumed {} and Republished {} messages. in {}".format(
                len(consumed_msgs),
                len(filtered_msgs),
                end_time - start_time))

        except Exception as e:
            LOGGER.critical("Could not republish, {}".format(e.message))

        time.sleep(config["interval"])


if __name__ == "__main__":

    parser = argparse.ArgumentParser(description="Republish messages for specific SITES")

    parser.add_argument(
        "-c", "--ConfigPath", type=str, help="Path for the config file", default="/etc/argo-messaging/republisher.json")

    parser.add_argument(
        "--verify", help="SSL verification for requests", dest="verify", action="store_true")

    parser.add_argument(
        "--debug", help="DEBUG mode", dest="debug", action="store_true")

    sys.exit(main(parser.parse_args()))
