# data bridge between MQTT-broker and Kafka 

```
go mod init github.com/pc/mqtt-bridge
go mod tidy
go build -o mqtt-bridge cmd/main.go
./mqtt-bridge
```

TODO
* MQTT subscribe (Done)
* MQTT autoReconnection (Done)
* MQTT subscribe auto recover
* MQTT publish

* publish to Kafka
* Kafka subscribe (autocommit or manualcommit)
