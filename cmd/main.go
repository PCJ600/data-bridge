package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/shr/cloud-agent/internal/mqtt"
	"github.com/shr/cloud-agent/internal/mq-adapter"
	"github.com/shr/cloud-agent/internal/msg-handler"
	"github.com/shr/cloud-agent/internal/handler"
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
)

func InitMQClient() (*mqttclient.MqttClient, *mqclient.MQClient){
	// 1. Create MQTT and MQ client instance.
    mqttBroker := os.Getenv("MQTT_BROKER")
    if mqttBroker == "" {
        mqttBroker = "tcp://emqx:1883"
    }
    mqttClientID := os.Getenv("MQTT_CLIENT_ID")
    if mqttClientID == "" {
        mqttClientID = fmt.Sprintf("cloud-agent-%s", uuid.Must(uuid.NewV7()))
    }
    mqBroker := os.Getenv("MQ_BROKER")
    if mqBroker == "" {
        mqBroker = "kafka:9092"
    }
    mqConsumerGroup := os.Getenv("MQ_CONSUMER_GROUP")
    if mqConsumerGroup == "" {
        mqConsumerGroup = "cloud-agent"
    }
	log.Printf("mqtt-broker: %s, mqtt-client: %s", mqttBroker, mqttClientID)
	log.Printf("mq-broker: %s, mq-consumer-group: %s", mqBroker, mqConsumerGroup)

	mqttClient := mqttclient.New(mqttBroker, mqttClientID)
	mqClient := mqclient.New(mqBroker, mqConsumerGroup)

	// 2. Register subscriptions for MQTT and MQ.
	h := msghandler.New(mqttClient, mqClient)
	mqttClient.RegisterSubscription("$share/mqtt_bridge/cloud/+/notify", h.EdgeGatewayMsgHandler)
	mqttClient.RegisterSubscription("$share/mqtt_bridge/cloud/+/telemetry", h.EdgeGatewayTelemetryMsgHandler)
	mqttClient.RegisterSubscription("$share/mqtt_bridge/cloud/+/model", h.EdgeGatewayModelMsgHandler)
	mqttClient.RegisterSubscription("$share/mqtt_bridge/cloud/+/alert", h.EdgeGatewayAlertMsgHandler)
	mqClient.RegisterSubscription("egw.notify", h.CloudMsgHandler)

	// 3. All dependencies ready, start MQTT and MQ client.
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

    demoTest := os.Getenv("DEMO_TEST")
	if demoTest == "true" {
		go MqttTest(mqttClient)
		go MQTest(mqClient)
	}

	// Start Gin
    ginMode := os.Getenv("GIN_MODE")
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(
		gin.Recovery(),
	)

	r.GET("/healthz", handler.HealthCheck)

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
