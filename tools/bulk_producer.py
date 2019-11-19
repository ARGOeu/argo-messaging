#!/usr/bin/env python

import json
import argo_ams_library
import requests
import argparse
import multiprocessing
import itertools
import os
import sys
import base64

ARGS = argparse.Namespace

PROJECT_NAME_FORMAT = "test_bulk_push_project_{}"
TOPIC_NAME_FORMAT = "test_bulk_push_topic_{}"
SUBSCRIPTION_NAME_FORMAT = "{}_sub_{}"


def create_projects():
    """
    Creates a pre defined number of projects specified by the arguments.

    Returns:
        projects: a list of strings, that represent projects' names .
    """

    projects = []

    for i in range(ARGS.number_of_projects):

        project_name = PROJECT_NAME_FORMAT.format(i)

        req_daa = {'description': project_name}

        url = "https://{}:{}/v1/projects/{}?key={}".format(ARGS.host, ARGS.port, project_name, ARGS.token)

        res = requests.post(url=url, data=json.dumps(req_daa), verify=ARGS.verify)

        print res.text

        projects.append(project_name)

    return projects


def create_topics(ams):
    """
    Creates a predefined number of topics specified by the arguments.

    Args:
        ams(ArgoMessagingService): an ams object providing the functionality we need to publish messages.

    Returns:
         topics: a list containing all the topics' names.
    """

    topics = []

    for i in range(ARGS.number_of_topics):

        topic_name = TOPIC_NAME_FORMAT.format(i)

        try:

            r = ams.create_topic(topic_name, verify=ARGS.verify)

            print r

        except Exception as e:

            print e.message

        topics.append(topic_name)

    return topics


def create_subs(topic, ams):
    """
    Creates a predefined number of subscriptions specified by the arguments.

    Args:
        topic(str): the topic's name that the subscription will be associated with
        ams(ArgoMessagingService): an ams object providing the functionality we need to publish messages.
    """

    for i in range(ARGS.number_of_subscriptions):

        sub_name = SUBSCRIPTION_NAME_FORMAT.format(topic, i)

        try:

            r = ams.create_sub(sub_name, topic, push_endpoint=ARGS.push_endpoint, verify=False)

            print r

        except Exception as e:

            print e.message


def publish(topic, ams):
    """Publish publishes a number of messages to an ams topic. The number of messages is determined by the arguments
        passed, as well as the size of each message.Each message is a collection of random bytes up the specified limit.

        Args:
           topic(str): the topic's name where the message will be published to.
           ams(ArgoMessagingService): an ams object providing the functionality we need to publish messages .
    """

    msgs = []

    for i in range(ARGS.number_of_messages):

        data = base64.b64encode(os.urandom(ARGS.message_size))

        msgs.append(argo_ams_library.AmsMessage(data=data))

    r = ams.publish(topic=topic, msg=msgs, verify=ARGS.verify)

    print r


def _publish(a_b):
    """
    Unpacks the arguments into the publish function. It is used as a helper.
    """
    return publish(*a_b)


def main():

    projects = create_projects()

    ams_endpoint = "{}:{}".format(ARGS.host, ARGS.port)

    for project in projects:

        ams = argo_ams_library.ArgoMessagingService(endpoint=ams_endpoint, token=ARGS.token, project=project)

        topics = create_topics(ams)

        for topic in topics:

            create_subs(topic, ams)

        pool = multiprocessing.Pool()
        pool.map(_publish, itertools.izip(topics, itertools.repeat(ams)))


if __name__ == "__main__":

    parser = argparse.ArgumentParser(description="Create projects/topics/subs to an AMS endpoint and publish messages")

    parser.add_argument(
        "-host", "--host", metavar="STRING", help="Ams endpoint",  type=str, dest="host", required=True)

    parser.add_argument(
        "-port", "--port", metavar="INTEGER", help="Ams port", default=443, type=int, dest="port")

    parser.add_argument(
        "-token", "--token", metavar="STRING", help="Ams token", type=str, dest="token", required=True)

    parser.add_argument(
        "-pe", "--push-endpoint", metavar="STRING", help="Subscriptions' push endpoint", type=str, default="", dest="push_endpoint")

    parser.add_argument(
        "-pn", "--projects-number", metavar="INTEGER", help="Number of ams projects",  default=1, type=int, dest="number_of_projects")

    parser.add_argument(
        "-tn", "--topics-number", metavar="INTEGER", help="Number of ams topics per project",  default=4, type=int, dest="number_of_topics")

    parser.add_argument(
        "-sn", "--subscriptions-number", metavar="INTEGER", help="Number of ams subscriptions per ams topic",  default=4, type=int, dest="number_of_subscriptions")

    parser.add_argument(
        "-mn", "--messages-number", metavar="INTEGER", help="Number of ams messages to be published to each ams topic", default=500, type=int, dest="number_of_messages")

    parser.add_argument(
        "-ms", "--message-size", metavar="INTEGER", help="The size of each message in bytes", default=1024, type=int, dest="message_size")

    parser.add_argument(
        "-v", "--verify", help="SSL verification for requests", dest="verify", action="store_true")

    ARGS = parser.parse_args()

    sys.exit(main())
