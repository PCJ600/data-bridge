package handler

import (
	"log"
	"encoding/json"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)


func EdgeGatewayHandler(_ mqtt.Client, msg mqtt.Message) {
    parts := strings.Split(msg.Topic(), "/")
    if len(parts) != 3 || parts[0] != "cloud" || parts[2] != "notify" {
        log.Printf("Invalid notify topic: %s", msg.Topic())
        return
    }
    edgeGatewayID := parts[1]

	payload := msg.Payload()
    var kafkaTopic string
    var data struct {
        Type string `json:"msgType"`
    }
    if err := json.Unmarshal(payload, &data); err == nil {
        switch data.Type {
        case "model":
            kafkaTopic = "cloud.model"
        case "alert":
            kafkaTopic = "cloud.alert"
        default:
			log.Printf("Unknown MQTT topic: %s, drop msg", msg.Topic())
        }
    } else {
		log.Printf("Can't parse msg, topic: %s, drop msg", payload, msg.Topic())
		return
    }

	// TODO: add Kafka Message to Chan
	log.Printf("Send msg to Kafka: Topic=%s, Key: %s", kafkaTopic, edgeGatewayID)
}

func EdgeGatewayTelemetryHandler(_ mqtt.Client, msg mqtt.Message) {
	log.Printf("Receive Telemetry Response: Topic=%s", msg.Topic())
}

func EdgeGatewayRespHandler(_ mqtt.Client, msg mqtt.Message) {
	log.Printf("Receive edge-gateway Response: Topic=%s", msg.Topic())
}
