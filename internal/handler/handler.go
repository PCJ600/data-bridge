package handler

import (
	"log"
	"encoding/json"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func EdgeGatewayCommonMsgHandler(_ mqtt.Client, msg mqtt.Message) {
    edgeGatewayID := strings.Split(msg.Topic(), "/")[1]

	payload := msg.Payload()
    var MQTopic string
    var data struct {
        Type string `json:"msgType"`
    }
    if err := json.Unmarshal(payload, &data); err == nil {
        switch data.Type {
        case "ack":
            MQTopic = "cloud.ack"
        default:
			log.Printf("Unknown MQTT topic: %s, drop msg", msg.Topic())
        }
    } else {
		log.Printf("parse msg error, topic: %s, drop msg", payload, msg.Topic())
		return
    }

	// TODO: publish to Kafka
	log.Printf("Send msg to Kafka: Topic=%s, Key: %s", MQTopic, edgeGatewayID)
}

func EdgeGatewayTelemetryMsgHandler(_ mqtt.Client, msg mqtt.Message) {
    edgeGatewayID := strings.Split(msg.Topic(), "/")[1]

	// TODO: publish to Kafka
	log.Printf("Send msg to Kafka: Topic=%s, Key: %s", msg.Topic(), edgeGatewayID)
}

func EdgeGatewayModelMsgHandler(_ mqtt.Client, msg mqtt.Message) {
    edgeGatewayID := strings.Split(msg.Topic(), "/")[1]

	// TODO: publish to Kafka
	log.Printf("Send msg to Kafka: Topic=%s, Key: %s", msg.Topic(), edgeGatewayID)
}

func EdgeGatewayAlertMsgHandler(_ mqtt.Client, msg mqtt.Message) {
    edgeGatewayID := strings.Split(msg.Topic(), "/")[1]

	// TODO: publish to Kafka
	log.Printf("Send msg to Kafka: Topic=%s, Key: %s", msg.Topic(), edgeGatewayID)
}


func CloudMsgHandler(topic string, key []byte, value []byte) {
	log.Printf("Receive msg from Kafka: Topic=%s, Key: %s", topic, key)
}

func MQSubscribeTestHandler(topic string, key []byte, value []byte) {
	log.Printf("Receive msg from Kafka: Topic=%s, Key: %s", topic, key)
}
