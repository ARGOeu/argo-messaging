#!/usr/bin/env python
from kafka import SimpleClient
from kafka.protocol.offset import OffsetRequest, OffsetResetStrategy
from kafka.common import OffsetRequestPayload
from kafka import KafkaClient
from kafka import KafkaConsumer
from kafka import TopicPartition
from kafka import KafkaProducer
import json
from pymongo import MongoClient
from argparse import ArgumentParser
import sys
import os
from glob import glob
import logging

logging.basicConfig(level=logging.INFO, format=logging.BASIC_FORMAT)
log = logging.getLogger('ams-migrate')

def get_mongo_db(mongo_host, mongo_port):
    """Get MongoDB database connection object
    
    Args:
        mongo_host (str): Mongodb server hostname
        mongo_port (int): Mongodb server port
    
    Returns:
        obj: MongoDB database connection object        
    """

    return MongoClient(mongo_host, mongo_port)["argo_msg"]

def get_kafka_client(broker_list):
    """Return a kafka client
    
    Args:
        broker_list (str): A comma separated list of kafka broker host:port values
    
    Returns:
        obj: kafka client  object
    """

    return SimpleClient(broker_list)


def get_kafka_producer(broker_list, batch_size=300):
    """Return a kafka producer
    
    Args:
        broker_list (str): A comma separated list of kafka broker host:port values
        batch_size (int, optional): Defaults to 300. Size of batch buffer to collect messages 
        before eachh send
    
    Returns:
        obj: A kafka producer object 
    """

    return KafkaProducer(
        bootstrap_servers=broker_list,
        batch_size=int(batch_size)
    )


def get_kafka_consumer(broker_list,timeout):
    """Return a kafka consumer
    
    Args:
        broker_list (str): A comma separated list of kafka broker host:port values
    
    Returns:
        obj: A kafka consumer object
    """

    return KafkaConsumer(
        group_id='ams-export',
        bootstrap_servers=broker_list,
        auto_offset_reset='smallest',
	    consumer_timeout_ms=500
    )
    

def get_mongo_topics(mongo_db):
    """Return a set with the names of AMS topics found in MongoDB
    
    Args:
        mongo_db (obj): MongoDB database connection object

    Returns:
        set(str): A set of strings (topic names)
    """     

    result = set()
    topics = mongo_db["topics"]
    for topic in topics.find():
        topic_name = "".join([topic["project_uuid"],".",topic["name"]])
        result.add(topic_name)
    
    log.info("Found %s topics in mongodb", len(result))
    return result

def get_kafka_topics(consumer):
    """Return a set with the name of topics found in kafka backend
    
    Args:
        consumer (obj): Kafka consumer object
    
    Returns:
        set(str): A set of strings (topic names)
    """

    return consumer.topics()

def get_actual_topics(mongo_db, k_consumer):
    """Return a set with the intersection of topics reported both 
    by MongoDB and Kafka backend
    
    Args:
        mongo_db (obj): MongoDB database connection object
        k_consumer (obj): Kafka consumer object
    
    Returns:
        set(str):  A set of strings (topic names)
    """

    m_topics = get_mongo_topics(mongo_db) 
    k_topics = get_kafka_topics(k_consumer)
    return m_topics.intersection(k_topics)

def get_topic_max(topic, k_client):
    """Return the max offset of a kafka topic
    
    Args:
        topic (str): Name of kafka topic
        k_client (obj): Kafka client object
    
    Returns:
        int: Max offset
    """

    partitions = k_client.topic_partitions[topic]
    offset_requests = [OffsetRequestPayload(topic, p, -1, 1) for p in partitions.keys()]
    offsets_responses = k_client.send_offset_request(offset_requests)

    for r in offsets_responses:
        if r.partition == 0:
            return r.offsets[0]

def export_topic(output, topic, k_consumer, max_off):
    """Fetch all available messages from a topic and 
    write the data in a text file
    
    Args:
        output (str): path to export the data to
        topic (str): Name of kafka topic
        k_consumer (obj): Kafka consumer object
        max (int): Maximum offset of topic
    """
    first_msg = True
    export_filename = output+"/"+topic+".topic.data"
    log.info("saving to: " + export_filename)
    with open(export_filename, "w") as topic_file:
        partition = TopicPartition(topic, 0)
        k_consumer.assign([partition])
        k_consumer.seek(partition, 0)
        for message in k_consumer:
            # If on first message, write min-max offsets
            # on the first line of export file as "[topic_name],[min],[max]\n"
            if first_msg:
                min_off=message.offset
                topic_file.write(topic+","+str(min_off)+","+str(max_off))
                topic_file.write("\n")
                first_msg = False
            json_txt = json.dumps(json.loads(message.value))
            topic_file.write(json_txt)
            topic_file.write("\n")
        if first_msg is True:
            # Write only header with min,max=max
            topic_file.write(topic+","+str(max_off)+","+str(max_off))
            topic_file.write("\n")



def export_data(args):
    """Main export routine
    
    Args:
        args (obj): command line arguments
    """
    log.info("Exporting ams data")
    mongo_args = args.mongo.split(":")
    log.info("Connect to mongo: " + args.mongo)
    mdb = get_mongo_db(mongo_args[0], int(mongo_args[1]))
    broker_args = args.brokers.split(",")
    log.info("Connect to kafka: " + args.brokers)
    k_consumer = get_kafka_consumer(broker_args, args.timeout)
    k_client = get_kafka_client(broker_args)
    topics = get_actual_topics(mdb, k_consumer)
    for topic in topics:
        max_off = get_topic_max(topic, k_client)
        log.info("exporting... " + topic+": "+str(max_off))
        export_topic(args.data, topic, k_consumer, max_off)





def main(args):
    """Connect to a kafka and MongoDB backend and 
    automatically imports or exports all AMS related topic data
    
    Args:
        args (obj): Command line arguments
    """
    log.info("hello")
    if args.cmd == 'export':
        export_data(args)
    elif args.cmd == 'import':
        log.info("import not yet implemented!")


    

if __name__ == "__main__":
    arg_parser = ArgumentParser(description="import/export ams data")
    
    arg_parser.add_argument(
        "--mongo", help="ams mongo host", dest="mongo", default="localhost:27017", metavar="string")
    arg_parser.add_argument(
        "--brokers", help="kafka broker list", dest="brokers", default="localhost:9092", metavar="string")
    arg_parser.add_argument(
        "--timeout", help="broker consume timeout (ms)", dest="timeout", default=300, metavar="int")
    arg_parser.add_argument(
        "--data", help="path where to export/import data", dest="data", default="./", metavar="int")
    arg_parser.add_argument(
        "--advance-offset", help="when importing advance offset to be aligned with original cluster", dest="advance", default=False, metavar="bool")
    arg_parser.add_argument(
        "--batch", help="how many messages per batch should be published", dest="batch_size", default=1000, metavar="int")
    cmd = arg_parser.add_subparsers(dest="cmd")
    cmd.add_parser('import', help='import ams data to kafka')
    cmd.add_parser('export', help='export ams data from kafka')
    # Parse the command line arguments accordingly and introduce them to
    # main...
    sys.exit(main(arg_parser.parse_args()))

