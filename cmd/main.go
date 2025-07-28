package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pc/mqtt-bridge/internal/mqtt"
	"github.com/pc/mqtt-bridge/internal/kafka"
	"github.com/pc/mqtt-bridge/internal/handler"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)


// TODO: get MQTT and Kafka broker from config
func InitMqttClient() (*mqttclient.MqttClient) {
	topic_handlers := map[string]mqtt.MessageHandler{
		"$share/mqtt_bridge/cloud/+/telemetry": handler.EdgeGatewayTelemetryHandler,
		"$share/mqtt_bridge/cloud/+/notify": handler.EdgeGatewayHandler,
		"$share/mqtt_bridge/cloud/+/ack": handler.EdgeGatewayRespHandler,
	}
	c := mqttclient.NewMqttClient("tcp://emqx:1883", topic_handlers)
	c.Init()
	return c
}

func InitKafkaClient() (*kafkaclient.KafkaClient) {
	c := kafkaclient.NewKafkaClient(
		"kafka:9092",
		"mqtt-bridge-consumer-group",
		[]string{"egw.ack", "egw.notify"},
	)
	c.Init()
	return c
}


func main() {
	InitMqttClient()
	client := InitKafkaClient()

	for {
		err := client.Publish("cloud.telemetry", "device-001", []byte(`{"msgType":"telemetry","temp":25.5}`))
		if err != nil {
			log.Printf("Publish error: %v", err)
		}
		time.Sleep(5 * time.Second)
	}


	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
