AMS data migration tools
========================

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

In a node with network access to both the AMS kafka backend and AMS mongo instance issue issue the following:

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