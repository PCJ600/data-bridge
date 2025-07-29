# data bridge between MQTT-broker and Kafka 

```
go mod init github.com/pc/mqtt-bridge
go mod tidy
go build -o mqtt-bridge cmd/main.go
./mqtt-bridge
```

Done
* MQTT pub/sub
* MQTT autoReconnect

* Kafka pub/sub
* Kafka autoReconnect
* Kafka subscriptions auto recover

TODO
* cpu/memory check
* print Kakfa Subscribe Logs
* test message not dropped
* MQTT subscriptions auto recover (?)
