package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/pc/mqtt-bridge/internal/mqtt"
	"github.com/pc/mqtt-bridge/internal/mq-adapter"
	"github.com/pc/mqtt-bridge/internal/msg-handler"
	"github.com/pc/mqtt-bridge/internal/handler"
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
)

func InitMQClient() (*mqttclient.MqttClient, *mqclient.MQClient){
	// Create MQTT and MQ client instance.
	mqttBroker := "tcp://emqx:1883"
	mqttClientID := fmt.Sprintf("mqtt-bridge-%s", uuid.Must(uuid.NewV7()))
	mqttClient := mqttclient.New(mqttBroker, mqttClientID)

	mqBroker := "kafka:9092"
	mqConsumerGroup := "mqtt-bridge"
	mqClient := mqclient.New(mqBroker, mqConsumerGroup)

	// Register subscriptions for MQTT and MQ.
	h := msghandler.New(mqttClient, mqClient)
	mqttClient.RegisterSubscription("$share/mqtt_bridge/cloud/+/notify", h.EdgeGatewayMsgHandler)
	mqttClient.RegisterSubscription("$share/mqtt_bridge/cloud/+/telemetry", h.EdgeGatewayTelemetryMsgHandler)
	mqttClient.RegisterSubscription("$share/mqtt_bridge/cloud/+/model", h.EdgeGatewayModelMsgHandler)
	mqttClient.RegisterSubscription("$share/mqtt_bridge/cloud/+/alert", h.EdgeGatewayAlertMsgHandler)
	mqClient.RegisterSubscription("egw.notify", h.CloudMsgHandler)

	// All dependencies ready, start MQTT and MQ client.
	mqttClient.Start()
	mqClient.Start()
	return mqttClient, mqClient
}


// Simulator MQTT Publisher
func MqttTest(c *mqttclient.MqttClient) {
    msgID := 0
    for {
        for deviceID := 1; deviceID <= 2; deviceID++ {
            topics := []string{"telemetry", "model", "alert", "notify"}
            for _, topic := range topics {
                fullTopic := fmt.Sprintf("cloud/%d/%s", deviceID, topic)
                payload := fmt.Sprintf(`{"msgType": "%s", "msgID": %d}`, topic, msgID)

                if err := c.Publish(fullTopic, []byte(payload)); err != nil {
                    log.Printf("Publish to MQ failed, msgType: %s, msgID: %d, err: %v", topic, msgID, err)
                }
                msgID++
            }
        }
        time.Sleep(5 * time.Second)
    }
}

// Simulator Kafka Publisher
func MQTest(c *mqclient.MQClient) {
    msgID := 0
	for {
        for deviceID := 1; deviceID <= 2; deviceID++ {
            payload := fmt.Sprintf(`{"msgType": "notify", "msgID": %d}`, msgID)
			err := c.Publish("egw.notify", []byte(strconv.Itoa(deviceID)), []byte(payload))
			if err != nil {
				log.Printf("Publish to MQTT failed, topic: %s, msgID: %d, err: %v", "egw.notify", msgID, err)
			}
			msgID++
		}
		time.Sleep(5 * time.Second)
	}
}



func main() {
	mqttClient, mqClient := InitMQClient()

	go MqttTest(mqttClient)
	go MQTest(mqClient)

	// Start Gin
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(
		gin.Recovery(),
	)

	r.GET("/healthz", handler.HealthCheck)

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
