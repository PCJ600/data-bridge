# data bridge

Persist MQTT messages and distribute them to cloud applications via Kakfa,
to achieve highly reliable transmission

# How to build
```
go mod init github.com/pc/mqtt-bridge
go mod tidy
go build -o mqtt-bridge cmd/main.go
./mqtt-bridge
```

# Run in docker
```

```

# Done
* MQTT pub/sub
* Kafka pub/sub
* MQTT autoReconnect
* Kafka autoReconnect
* Data Bridge handler

# TODO
* Docker compose
* EMQX and Kafka Auth
* Access EMQX and Kafka cluster using load balance
* Benchmark test, cpu, memory check, test message not dropped
