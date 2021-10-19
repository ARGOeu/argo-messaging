#!/usr/bin/env python

from kazoo.client import KazooClient
import json
import argparse
import sys
import subprocess

TOPICS_PATH = "/brokers/topics/"
REASSIGNMENT_SCRIPT = "/kafka-reassign-partitions.sh"

def get_zookeeper_topics_partitions(zoo_list):
    """
    Get topics partitions from zookeeper

    Args:
        zoo_list (str): A comma separated list of zookeeper hosts and ports, e.g. zoo1:2181,zoo2:2181,zoo3:2181
    Returns:
        list: A list that each entry represents a topic with its partitions
    """

    topics_partitions = []

    zk = KazooClient(hosts=zoo_list)
    zk.start()
    for zoo_kafka_topic in zk.get_children("/brokers/topics"):
        zoo_kafka_topic_path = TOPICS_PATH + zoo_kafka_topic
        partitions = json.loads((zk.get(zoo_kafka_topic_path)[0]).decode('utf-8'))['partitions']
        topic = {
            "topic_name": zoo_kafka_topic,
            "partitions": partitions
        }

        topics_partitions.append(topic)
    zk.stop()
    return topics_partitions

def generate_conf(topics_partitions, broker_remove, broker_add, replication_factor):
    """
    Generate the needed config for the reassignment tool and rollback process

    Args:
        topics_partitions (list): List that each entry represents a topic with its partitions
        broker_remove (int): Broker id that is being decommissioned
        broker_add (int): Broker id that is being added
        replication_factor(int): Number of ISR each partition should have
    Returns:
        dict: Dict that contains two entries, the "rollback" entry which represents the state of the cluster before
               the reassignment AND "reassignment" which contains the configuration needed for the reassignment tool.
    """
    reassignment_partitions_conf = []
    reassignment_conf = {"version": 1, "partitions": []}

    rollback_partitions_conf = []
    rollback_conf = {"version": 1, "partitions": []}

    leader_reassign_count = 0
    follower_reassign_count = 0

    for topic_partitions in topics_partitions:
        for partition_number in topic_partitions["partitions"]:

            topic_name = topic_partitions["topic_name"]

            print("Examining topic: {0} partition {1} . . .".format(topic_name, partition_number))

            in_sync_replicas = topic_partitions["partitions"][partition_number]

            # set up the another rollback entry for the specific topic/partition
            rollback_topic_dict = {
                "topic": topic_name,
                "partition": int(partition_number),
                "replicas": in_sync_replicas
            }
            rollback_partitions_conf.append(rollback_topic_dict)


            # if the broker id in found in the ISR modify the entry
            if broker_remove in in_sync_replicas:
                reassigned_in_sync_replicas = in_sync_replicas[:]
                i = list(reassigned_in_sync_replicas).index(broker_remove)
                if i == 0:
                    # the first entry in the isr list is the leader
                    leader_reassign_count += 1
                    print("Replacing leader for topic {0} and partition {1}"
                          .format(topic_name, partition_number))
                else:
                    follower_reassign_count += 1
                    print("Replacing follower for topic {0} and partition {1}"
                          .format(topic_name, partition_number))

                reassigned_in_sync_replicas[i] = broker_add

                if len(reassigned_in_sync_replicas) > replication_factor:
                    print("Found more ISR {0} than replication factor for topic: {1} and partition {2}"
                          .format(reassigned_in_sync_replicas, topic_name, partition_number))
                    reassigned_in_sync_replicas = reassigned_in_sync_replicas[:replication_factor]

                reassign_topic_dict = {
                    "topic": topic_name,
                    "partition": int(partition_number),
                    "replicas": reassigned_in_sync_replicas
                }

                reassignment_partitions_conf.append(reassign_topic_dict)
            print("- - - - - - - - - - - - - - ")

    reassignment_conf["partitions"] = reassignment_partitions_conf
    rollback_conf["partitions"] = rollback_partitions_conf

    print("Total leader reassignments: {0}".format(leader_reassign_count))
    print("Total follower reassignments: {0}".format(follower_reassign_count))

    return {
        "rollback": rollback_conf,
        "reassignment": reassignment_conf
    }

def build_reassignment_command(zoo_list, kafka_bin_dir, re_json_file):
    """
    Build the command that executes the kafka-reassign-partitions.sh script with the appropriate arguments

    Args:
        zoo_list (str): A comma separated list of zookeeper hosts and ports, e.g. zoo1:2181,zoo2:2181,zoo3:2181
        kafka_bin_dir (str): Kafka installation bin directory where all the administrative tools are being kept
        re_json_file (str): File location for the reassignment-json-file
    Returns:
        list: A list that contains the main command and its arguments
    """

    cmd = list()

    # append the call to the reassignment script
    cmd.append(kafka_bin_dir + REASSIGNMENT_SCRIPT)

    # zookeeper hosts list
    cmd.append("--zookeeper")
    cmd.append(zoo_list)

    # reassignment json file
    cmd.append("--reassignment-json-file")
    cmd.append(re_json_file)

    # execute command
    cmd.append("--execute")

    return cmd

def main(args):

    topics_partitions = get_zookeeper_topics_partitions(args.zoo_list)

    conf = generate_conf(topics_partitions, args.broker_remove, args.broker_add, args.replication_factor)

    # save the reassignment conf
    with open(args.reassignment_file, 'w', encoding='utf-8') as f:
        json.dump(conf["reassignment"], f, ensure_ascii=False, indent=4)

    # save the rollback conf
    with open(args.rollback_file, 'w', encoding='utf-8') as f:
        json.dump(conf["rollback"], f, ensure_ascii=False, indent=4)


    if args.execute:

        reassignment_command = build_reassignment_command(args.zoo_list, args.kafka_bin_dir, args.reassignment_file)

        print("Executing command: {0}".format(" ".join(reassignment_command)))

        print(subprocess.check_call(reassignment_command))


if __name__ == '__main__':

    parser = argparse.ArgumentParser(description="Reassign partitions from one broker to another")

    parser.add_argument(
        "--zoo-list", type=str, help="Comma separated list of zookeeper hosts, host1:port1,host1:port2,host3:port3")

    parser.add_argument(
        "--broker-remove", type=int, help="The id of the broker that is being decommissioned")

    parser.add_argument(
        "--broker-add", type=int, help="The id of the broker that is being added to the cluster")


    parser.add_argument(
        "--replication-factor", type=int, default=2, help="The number of ISR each partition should have including the leader")

    parser.add_argument(
        "--kafka-bin-dir", type=str,
        help="Kafka installation bin directory where all the administrative tools are being kept."
        "The script will call --kafka-bin-dir/kafka-reassign-partitions.sh",
        default="/usr/lib/kafka/bin"
    )

    parser.add_argument(
        "--rollback-file", type=str,
        help="Location for the rollback file that will be generated.The rollback file holds the cluster state before the reassignment",
        default="rollback.json"
    )

    parser.add_argument(
        "--reassignment-file", type=str,
        help="Location for the reassignment file that will be generated."
             "The file will later be used an an input to kafka-reassign-partitions.sh.",
        default="reassignment.json"
    )

    parser.add_argument(
        "--execute", help="Execute will call kafka's reassign script and begin the process", dest="execute", action="store_true")

    sys.exit(main(parser.parse_args()))