AMS data tools
========================

stream_producer
----------------
Stream producer is a script that allows you to connect to an AMS endpoint and publish messages of configurable size indefinitely.

Requirements
------------

- argo_ams_library

How to run stream_producer
--------------------------

`./stream_producer.py --host some.ams.host --port 443 --token some_ams_token --project ams_project --topic ams_topic
--bulk-size 10 --message-size 4096 --fire-rate 5`

- `-host, --host` is the AMS endpoint to connect to.
- `-port, --port` is the AMS port.
- `-token, --token` is the AMS token that will grant you access to perform all the needed actions.
- `-project, --project` is the AMS project that the topic belongs to.
- `topic, --topic` is the AMS topic that the messages will be published to.
- `-bs, --bulk-size` is the amount of messages to publish to each topic in every request, `default=1`.
- `-ms, --message-size` is the size of each message in bytes, `default=1024`.
- `fr, --fire-rate` is the interval at which the messages will be published, `default=0`.
- `-v, --verify` whether or not to do ssl verification, `if left undeclared, it will not verify`.


bulk_producer
----------------
Bulk producer is a script that allows you to connect to an AMS endpoint and create projects/topics/subscriptions
and publish a predefined amount of messages in order to test the functionality of the service.

Requirements
------------

- requests
- argo_ams_library

How to run bulk_producer
------------------------

`./bulk_producer.py --host some.ams.host --port 443 --token some_ams_token --projects-number 2 --topics-number 4 --subscriptions-number 4
--messages-number 1000 --message-size 1024 --push-endpoint https://127.0.0.1:5000/receive_here --verify`

- `-host, --host` is the AMS endpoint to connect to.
- `-port, --port` is the AMS port.
- `-token, --token` is the AMS token that will grant you access to perform all the needed actions.
- `pn, --projects-number` is the number of AMS projects to be created, `default=1`.
- `tn, --topics-number` is the number of topics to create under each project, `default=4`.
- `sn, --subscriptions-number` is the amount of subscriptions to assign to each topic, `default=4`.
- `-mn, messages-number` is the amount of messages to publish to each topic, `default=500`.
- `-ms, --message-size` is the size of each message in bytes, `default=1024`.
- `pe, --push-endpoint` is the end where the subscriptions will push the messages they consume, `if left undeclared, the subscriptions will be in pull mode`.
- `-v, --verify` whether or not to do ssl verification, `if left undeclared, it will not verify`.

consumer
----------------
Consumer is a script that allows you to connect to an AMS endpoint and consume(pull) messages
 of configurable size indefinitely.

Requirements
------------

- argo_ams_library

How to run consumer
--------------------------

`./consumer.py --host some.ams.host --port 443 --token some_ams_token --project ams_project --sub sub-1
--bulk-size 10 --fire-rate 5`

- `-host, --host` is the AMS endpoint to connect to.
- `-port, --port` is the AMS port.
- `-token, --token` is the AMS token that will grant you access to perform all the needed actions.
- `-project, --project` is the AMS project that the topic belongs to.
- `sub, --sub` is the AMS sub that the messages will be consumed from.
- `-bs, --bulk-size` is the amount of messages to publish to each topic in every request, `default=1`.
- `fr, --fire-rate` is the interval at which the messages will be published, `default=0`.
- `-v, --verify` whether or not to do ssl verification, `if left undeclared, it will not verify`.

ams_kafka_export
----------------

This command line tool can be used to export/import data from AMS kafka topics into text files and move them to another AMS kafka cluster. 

Requirements
------------

To run the script you need python 2.7 and the following libraries:

- pymongo
- kafka-python  

How to run for export
---------------------

In a node with network access to both the AMS kafka backend and AMS mongo instance issue the following:

$ `./ams-migrate.py --mongo "localhost:27017" --brokers "localhost:9092" --timeout 300 --data ./ export`
or
$ `./ams-migrate.py export` filled with default values targeting localhost

where `--mongo` follow the hostname:port of mongodb 
where `--brokers` follow with a comma-separated list of host:port of kafka instances
where `--timeout` specify a consume wait timeout in milliseconds
where `--data` specify a folder to export data

How to run for import
---------------------

In a node with network access to both the AMS kafka backend and AMS mongo instance issue issue the following:

$ `./ams-migrate.py --mongo "localhost:27017" --brokers "localhost:9092" --batch 300 --advance-offsets false --data ./ import`
or
$ `./ams-migrate.py import` filled with default values targeting localhost

where `--mongo` follow the hostname:port of mongodb
where `--brokers` follow with a comma-separated list of host:port of kafka instances
where `--batch` specify num of messages per batch import operation
where `--data` specify a folder to import data from
where `--advance-offsets` if true, advance topic offsets by publishing empty messages

Exported file types
-------------------

topics are exported in disk as files with the following filename pattern:
`topic_name.topic.data`

Each exported topic file is a text file with one exported message per line.
First line is reserved as header with the following topic metadata:
`topic_name`,`minimum_offset`,`maximum_offset`

Example contents of `foo.topic.data`:

```text
foo,0,15
message1
message2
message3
```