# data bridge between MQTT-broker and Kafka 

```
go mod init github.com/pc/mqtt-bridge
go mod tidy
go build -o mqtt-bridge cmd/main.go
./mqtt-bridge
```

TODO
* test autoReconnection for MQTT
* feat: send to Kafka
