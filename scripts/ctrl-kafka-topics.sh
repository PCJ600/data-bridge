#!/bin/bash

kafka_topics_bin="bin/kafka-topics.sh"
kafkaserver="kafka:9092"
replica=1
parts=6
action="$1"
topic="$2"

function create_kafka_topics() {
    # 1. cloud.telemetry
    ${kafka_topics_bin} --create \
      --bootstrap-server ${kafkaserver} \
      --topic cloud.telemetry \
      --partitions ${parts} \
      --replication-factor ${replica} \
      --config retention.ms=3600000 \
      --config compression.type=lz4
    
    # 2. cloud.model
    ${kafka_topics_bin} --create \
      --bootstrap-server ${kafkaserver} \
      --topic cloud.model \
      --partitions ${parts} \
      --replication-factor ${replica} \
      --config retention.ms=604800000 \
      --config compression.type=lz4
    
    # 3. cloud.alert
    ${kafka_topics_bin} --create \
      --bootstrap-server ${kafkaserver} \
      --topic cloud.alert \
      --partitions ${parts} \
      --replication-factor ${replica} \
      --config retention.ms=604800000 \
      --config compression.type=lz4
    
    # 4. cloud.notify
    ${kafka_topics_bin} --create \
      --bootstrap-server ${kafkaserver} \
      --topic cloud.notify \
      --partitions ${parts} \
      --replication-factor ${replica} \
      --config retention.ms=3600000 \
      --config compression.type=lz4
    
    # 5. egw.notify
    ${kafka_topics_bin} --create \
      --bootstrap-server ${kafkaserver} \
      --topic egw.notify \
      --partitions ${parts} \
      --replication-factor ${replica} \
      --config retention.ms=3600000 \
      --config compression.type=lz4
}

function delete_kafka_topics() {
	local topic="$1"
	if [ ! -z ${topic} ]; then
		${kafka_topics_bin} --bootstrap-server ${kafkaserver} --delete --topic "${topic}"
	fi
}

function list_kafka_topics() {
	${kafka_topics_bin} --list --bootstrap-server ${kafkaserver}
}

if [ "${action}" == "delete" ]; then
	delete_kafka_topics ${topic}
elif [ "${action}" == "list" ]; then
	list_kafka_topics
elif [ "${action}" == "create" ]; then
	create_kafka_topics
else
	echo "no args given"
fi
