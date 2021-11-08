#!/usr/bin/env python

from kafka import KafkaConsumer
from kafka import KafkaAdminClient
from kafka.errors import UnknownTopicOrPartitionError
from pymongo import MongoClient
import sys
import argparse

def main(args):

    if args.dry:
        print("--- This a DRY run. No topic will be deleted! ---")

    # set up the bootstrap servers for kafka client
    bootstrap_servers = [x for x in args.broker_list.split(",")]

    kafka_consumer = KafkaConsumer(bootstrap_servers=bootstrap_servers)
    kafka_topics = kafka_consumer.topics()

    kafka_admin_client = KafkaAdminClient(bootstrap_servers=bootstrap_servers)

    topics_col = MongoClient(args.mongo_host, args.mongo_port).get_database(name="argo_msg").get_collection(name="topics")
    ams_topics = set()
    for top in topics_col.find():
        ams_topics.add("{0}.{1}".format(top["project_uuid"], top["name"]))

    topics_to_delete_count = 0
    topics_to_be_deleted = kafka_consumer.topics().difference(ams_topics)
    for top_to_del in topics_to_be_deleted:
        print("Marking topic: " + str(top_to_del) + " for deletion (X)")
        if not args.dry:
            try:
                print(kafka_admin_client.delete_topics(topics=[top_to_del]))
                topics_to_delete_count += 1
                print("---------------------------------------------------")
            except UnknownTopicOrPartitionError as e:
                print("Could not delete topic {0}. Exception: {1}"
                      .format(top_to_del, str(e.message)))
                print("---------------------------------------------------")
                continue

    print("Total Kafka topics: {0}".format(len(kafka_topics)))
    print("Total AMS topics: {0}".format(len(ams_topics)))
    print("Total Marked topics: {0}".format(len(topics_to_be_deleted)))
    print("Total Deleted topics: {0}".format(topics_to_delete_count))


if __name__ == '__main__':

    parser = argparse.ArgumentParser(description="Delete Kafka topics that are not present in AMS")

    parser.add_argument(
        "--broker-list", type=str, help="Comma separated list of brokers, host1:port1,host1:port2,host3:port3")

    parser.add_argument(
        "--mongo-host", type=str, help="Mongo host", default="127.0.0.1")

    parser.add_argument(
        "--mongo-port", type=int, help="Mongo port", default=27017)

    parser.add_argument(
        "--dry", help="DRY run", dest="dry", action="store_true")

    sys.exit(main(parser.parse_args()))