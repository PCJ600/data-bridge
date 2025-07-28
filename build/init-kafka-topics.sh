#!/bin/bash

kafkaserver="kafka:9092"
replica=1
parts=6

# 1. cloud.telemetry (保留1小时)
bin/kafka-topics.sh --create \
  --bootstrap-server ${kafkaserver} \
  --topic cloud.telemetry \
  --partitions ${parts} \
  --replication-factor ${replica} \
  --config retention.ms=3600000

# 2. cloud.model (保留7天)
bin/kafka-topics.sh --create \
  --bootstrap-server ${kafkaserver} \
  --topic cloud.model \
  --partitions ${parts} \
  --replication-factor ${replica} \
  --config retention.ms=604800000

# 3. cloud.alert (保留7天)
bin/kafka-topics.sh --create \
  --bootstrap-server ${kafkaserver} \
  --topic cloud.alert \
  --partitions ${parts} \
  --replication-factor ${replica} \
  --config retention.ms=604800000

# 4. egw.notify (保留1小时)
bin/kafka-topics.sh --create \
  --bootstrap-server ${kafkaserver} \
  --topic egw.notify \
  --partitions ${parts} \
  --replication-factor ${replica} \
  --config retention.ms=3600000

# 5. cloud.ack (保留1小时)
bin/kafka-topics.sh --create \
  --bootstrap-server ${kafkaserver} \
  --topic cloud.ack \
  --partitions ${parts} \
  --replication-factor ${replica} \
  --config retention.ms=3600000

# 6. egw.ack (保留1小时)
bin/kafka-topics.sh --create \
  --bootstrap-server ${kafkaserver} \
  --topic egw.ack \
  --partitions ${parts} \
  --replication-factor ${replica} \
  --config retention.ms=3600000


