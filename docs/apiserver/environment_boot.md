# About how to boot zookeeper and kafka:

Note: this might be integrated into a boot script later.

First boot zookeeper:
$ zookeeper-server-start.sh /usr/local/kafka/config/zookeeper.properties&

Be aware to add ampersand '&' to run this in background.

Then boot kafka:
$ kafka-server-start.sh /usr/local/kafka/config/server.properties&

Then you should be able to use kafka now!

# About how to close zookeeper and kafka:

$ kafka-server-stop.sh
$ zookeeper-server-stop.sh

Then run:
$ ps -a

You shall not find any process called 'java' running.