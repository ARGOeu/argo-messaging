#!/usr/bin/env python

import argparse
import argo_ams_library
import sys
import os
import base64
import time


def main(args):

    ams_endpoint = "{}:{}".format(args.host, args.port)

    ams = argo_ams_library.ArgoMessagingService(endpoint=ams_endpoint, token=args.token, project=args.project)

    while True:

        try:

            msgs = []

            for i in range(args.bulk_size):

                data = base64.b64encode(os.urandom(args.message_size))

                msgs.append(argo_ams_library.AmsMessage(data=data))

            r = ams.publish(topic=args.topic, msg=msgs, verify=args.verify)

            print r

            time.sleep(args.fire_rate)

        except Exception as e:
            print "Couldn't publish to topic {}, {}".format(args.topic, e.message)
            continue


if __name__ == "__main__":

    parser = argparse.ArgumentParser(description="Publish messages to an ams topic indefinitely")

    parser.add_argument(
        "-host", "--host", metavar="STRING", help="Ams endpoint", type=str, dest="host", required=True)

    parser.add_argument(
        "-port", "--port", metavar="INTEGER", help="Ams port", default=443, type=str, dest="port")

    parser.add_argument(
        "-token", "--token", metavar="STRING", help="Ams token", type=str, dest="token", required=True)

    parser.add_argument(
        "-project", "--project", metavar="STRING", help="Ams project", type=str, dest="project", required=True)

    parser.add_argument(
        "-topic", "--topic", metavar="STRING", help="Ams topic", type=str, dest="topic", required=True)

    parser.add_argument(
        "-bs", "--bulk-size", metavar="INTEGER", help="Number of ams messages to be published with each request", default=1, type=int, dest="bulk_size")

    parser.add_argument(
        "-ms", "--message-size", metavar="INTEGER", help="The size of each message in bytes", default=1024, type=int, dest="message_size")

    parser.add_argument(
        "-fr", "--fire-rate", metavar="INTEGER", type=int, default=0,
        help="How often should it try to publish in milliseconds", dest="fire_rate")

    parser.add_argument(
        "-v", "--verify", help="SSL verification for requests", dest="verify", action="store_true")

    sys.exit(main(parser.parse_args()))
