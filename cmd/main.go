package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pc/mqtt-bridge/internal/mqtt"
	"github.com/pc/mqtt-bridge/internal/mq-adapter"
	"github.com/pc/mqtt-bridge/internal/handler"

	"github.com/google/uuid"
)

/*
topic_handlers := map[string]mqclient.MessageHandler{
	"test.topic": handler.MQSubscribeTestHandler,
}

topic_handlers := map[string]mqtt.MessageHandler{
		"$share/mqtt_bridge/cloud/+/notify": handler.EdgeGatewayCommonMsgHandler,
		"$share/mqtt_bridge/cloud/+/model": handler.EdgeGatewayModelMsgHandler,
		"$share/mqtt_bridge/cloud/+/alert": handler.EdgeGatewayAlertMsgHandler,
	}
*/
func InitMQClient() (*mqttclient.MqttClient, *mqclient.MQClient){
	// 1. Create MQTT and MQ client instance.
	mqttBroker := "tcp://emqx:1883"
	mqttClientID := fmt.Sprintf("mqtt-bridge-%s", uuid.Must(uuid.NewV7()))
	mqttClient := mqttclient.New(mqttBroker, mqttClientID)

	mqBroker := "kafka:9092"
	mqConsumerGroup := "mqtt-bridge"
	mqClient := mqclient.New(mqBroker, mqConsumerGroup)

	// 2. Register subscriptions handlers for MQTT and MQ.
	h := handler.New(mqttClient, mqClient)
	mqttClient.RegisterSubscription("$share/mqtt_bridge/cloud/+/telemetry", h.EdgeGatewayTelemetryMsgHandler)
	mqClient.RegisterSubscription("egw.notify", h.CloudMsgHandler)

	// 3. All dependencies ready, start MQTT and MQ client.
	mqttClient.Start()
	mqClient.Start()
	return mqttClient, mqClient
}

func MqttTest(c *mqttclient.MqttClient) {
	for {
		err := c.Publish("cloud/1/telemetry", []byte(`{"temp": "25.5"}`))
		if err != nil {
			log.Printf("MQTT Publish error: %v", err)
		}
		time.Sleep(3 * time.Second)
	}
}

func MQTest(c *mqclient.MQClient) {
	for {
		err := c.Publish("egw.notify", []byte("1"), []byte(`{"msgType":"upgrade"}`))
		if err != nil {
			log.Printf("Publish error: %v", err)
		}
		time.Sleep(3 * time.Second)
	}
}

func main() {
	mqttClient, mqClient := InitMQClient()

	go MqttTest(mqttClient)
	go MQTest(mqClient)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
