package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pc/mqtt-bridge/internal/mqtt"
	"github.com/pc/mqtt-bridge/internal/handler"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)


func InitMqttClient() (*mqttclient.MqttClient){
	c := mqttclient.New()
	topic_handlers := map[string]mqtt.MessageHandler{
		"$share/mqtt_bridge/cloud/+/telemetry": handler.EdgeGatewayTelemetryHandler,
		"$share/mqtt_bridge/cloud/+/notify": handler.EdgeGatewayHandler,
		"$share/mqtt_bridge/cloud/+/ack": handler.EdgeGatewayRespHandler,
	}
	c.Init("tcp://emqx:1883", topic_handlers)
	return c
}


func main() {
	InitMqttClient()
	// defer c.Disconnect()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
