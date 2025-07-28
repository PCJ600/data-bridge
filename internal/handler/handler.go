package handler

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func EdgeGatewayHandler(_ mqtt.Client, msg mqtt.Message) {
	log.Printf("Receive edge-gateway data: Topic=%s", msg.Topic())
}

func EdgeGatewayRespHandler(_ mqtt.Client, msg mqtt.Message) {
	log.Printf("Receive edge-gateway Response: Topic=%s", msg.Topic())
}
