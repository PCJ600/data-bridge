package handler

import (
    "fmt"
    "log"
    "strings"

	"github.com/pc/mqtt-bridge/internal/mqtt"
	"github.com/pc/mqtt-bridge/internal/mq-adapter"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Handler handles bidirectional message bridging between MQTT and MQ.
type Handler struct {
    mqttPublisher mqttclient.Publisher
    mqPublisher   mqclient.Publisher
}

// NewHandler creates a new Handler with the required publishers.
func New(mqttPublisher mqttclient.Publisher, mqPublisher mqclient.Publisher) *Handler {
    return &Handler{
        mqttPublisher: mqttPublisher,
        mqPublisher:   mqPublisher,
    }
}

// Handles telemetry data from edge gateway (received via MQTT), and forwards to MQ for further processing.
func (h *Handler) EdgeGatewayTelemetryMsgHandler(_ mqtt.Client, msg mqtt.Message) {
    if h.mqPublisher == nil {
        log.Printf("MQ publisher not set")
        return
    }

    edgeGatewayID := strings.Split(msg.Topic(), "/")[1]
	// TODO: async publish
    err := h.mqPublisher.Publish("cloud.telemetry", []byte(edgeGatewayID), msg.Payload())
    if err != nil {
        log.Printf("Failed to send telemetry msg to MQ: %v", err)
        return
    }
    log.Printf("Sent telemetry msg to MQ: topic=%s", msg.Topic())
}

// Handles downstream message from MQ, and forwards to MQTT for edge gateway process.
func (h *Handler) CloudMsgHandler(topic string, key, value []byte) {
    if h.mqttPublisher == nil {
        log.Printf("MQTT publisher not set")
        return
    }

	mqttTopic := fmt.Sprintf("egw/%s/notify", string(key))
    err := h.mqttPublisher.Publish(mqttTopic, value)
    if err != nil {
        log.Printf("Failed to send msg to MQTT: %v", err)
        return
    }
    log.Printf("Sent msg to MQTT: topic=%s", topic)
}
