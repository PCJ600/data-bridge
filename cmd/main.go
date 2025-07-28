package main

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pc/mqtt-bridge/internal/mqtt"
	"github.com/pc/mqtt-bridge/internal/handler"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)


func InitMqttClient() (*mqttclient.MqttClient, error){
	c := mqttclient.New()
	topic_handlers := map[string]mqtt.MessageHandler{
		"cloud/+/notify": handler.EdgeGatewayHandler,
		"cloud/+/ack": handler.EdgeGatewayRespHandler,
	}
	c.Init("tcp://emqx:1883", topic_handlers)

	retryCount := 0
	maxRetries := 10
	for {
		err := c.Connect()
		if err == nil {
			return c, nil
		}
		retryCount++
		if retryCount >= maxRetries {
			log.Fatalf("Failed to connect MQTT after %d attempts: %v", maxRetries, err)
		}
		log.Printf("MQTT Connection attempt %d/%d failed: %v", retryCount, maxRetries, err)
		time.Sleep(5 * time.Second)
	}
	return nil, errors.New("Init MQTT Connect failed")
}


func main() {
	mqttc, err := InitMqttClient()
	if err != nil {
		panic(err.Error())
	}
	defer mqttc.Disconnect()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

}
