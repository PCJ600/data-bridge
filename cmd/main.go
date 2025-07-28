package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	// "time"

	"github.com/pc/mqtt-bridge/internal/mqtt"
	"github.com/pc/mqtt-bridge/internal/kafka"
	"github.com/pc/mqtt-bridge/internal/handler"

	"github.com/google/uuid"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// TODO: get MQTT and Kafka broker, auth user/password from config

func InitMqttClient() (*mqttclient.MqttClient) {
	broker := "tcp://emqx:1883"
	clientID := fmt.Sprintf("mqtt-bridge-%s", uuid.Must(uuid.NewV7()))
	topic_handlers := map[string]mqtt.MessageHandler{
		"$share/mqtt_bridge/cloud/+/notify": handler.EdgeGatewayCommonMsgHandler,
		"$share/mqtt_bridge/cloud/+/telemetry": handler.EdgeGatewayTelemetryMsgHandler,
		"$share/mqtt_bridge/cloud/+/model": handler.EdgeGatewayModelMsgHandler,
		"$share/mqtt_bridge/cloud/+/alert": handler.EdgeGatewayAlertMsgHandler,
	}
	c := mqttclient.NewMqttClient(broker, clientID, topic_handlers)
	c.Init()
	return c
}

func InitKafkaClient() (*kafkaclient.KafkaClient) {
	broker := "kafka:9092"
	consumerGroup := "mqtt-bridge"
	topic_handlers := map[string]kafkaclient.MessageHandler{
		"egw.notify": handler.CloudMsgHandler,
		"test.topic": handler.KafkaSubscribeTestHandler,
	}
	c := kafkaclient.NewKafkaClient(broker, consumerGroup, topic_handlers)
	c.Init()
	return c
}


func main() {
	mqttc := InitMqttClient()
	kafkac := InitKafkaClient()
	log.Printf("mqttclient: %v, kafkac: %v", mqttc, kafkac)

	// MQTT publish
	/*
	for {
		err := mqttc.Publish("cloud/12345/notify", []byte(`{"msgType": "ack","command":"reboot","delay":10}`))
		if err != nil {
			log.Printf("MQTT Publish error: %v", err)
		}
		time.Sleep(3 * time.Second)
	}

	for {
		err := kafkac.Publish("egw.notify", "device-001", []byte(`{"msgType":"upgrade"}`))
		if err != nil {
			log.Printf("Publish error: %v", err)
		}
		time.Sleep(3 * time.Second)
	}
	*/


	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
