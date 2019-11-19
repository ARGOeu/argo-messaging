#!/usr/bin/env python

from argo_ams_library import ArgoMessagingService
import argparse
import sys
import time


def main(args):

    ams = ArgoMessagingService(endpoint=args.host, token=args.token, project=args.project)

    while True:
        try:

            consumed_msgs = ams.pull_sub(sub=args.sub, num=args.bulk_size, return_immediately=True, verify=args.verify)

            last_msg_id = "-1"
            if len(consumed_msgs) > 0:
                last_msg_id = consumed_msgs.pop()[0]

            print last_msg_id
            print "\n"

            if last_msg_id != "-1":
                print ams.ack_sub(args.sub, [last_msg_id], verify=args.verify)

            time.sleep(args.fire_rate)

        except Exception as e:
            print "Couldn't consume from sub {}, {}".format(args.sub, e.message)
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
        "-sub", "--sub", metavar="STRING", help="Ams sub", type=str, dest="sub", required=True)

    parser.add_argument(
        "-bs", "--bulk-size", metavar="INTEGER", help="Number of ams messages to be consumed with each request", default=1, type=int, dest="bulk_size")

    parser.add_argument(
        "-fr", "--fire-rate", metavar="INTEGER", type=int, default=0,
        help="How often should it try to consume in milliseconds", dest="fire_rate")

    parser.add_argument(
        "-v", "--verify", help="SSL verification for requests", dest="verify", action="store_true")

    sys.exit(main(parser.parse_args()))
