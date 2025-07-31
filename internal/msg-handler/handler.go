package msghandler

import (
    "fmt"
    "log"
    "strings"
	"sync"

	"github.com/shr/cloud-agent/internal/mqtt"
	"github.com/shr/cloud-agent/internal/mq-adapter"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Handler handles bidirectional message bridging between MQTT and MQ.
type mqJob struct {
    topic   string
    payload []byte
    key     []byte
}

type mqttJob struct {
    topic   string
    payload []byte
}

type Handler struct {
    mqttPublisher mqttclient.Publisher
    mqPublisher   mqclient.Publisher

	mqttJobQueue  chan mqttJob
	mqJobQueue    chan mqJob
	running       bool
	wg            sync.WaitGroup
}

// NewHandler creates a new Handler with the required publishers.
func New(mqttPublisher mqttclient.Publisher, mqPublisher mqclient.Publisher) *Handler {
    handler := &Handler{
        mqttPublisher:    mqttPublisher,
        mqPublisher:      mqPublisher,
		mqttJobQueue:     make(chan mqttJob, 10000),
		mqJobQueue:       make(chan mqJob, 10000),
		running:          true,
    }

	handler.startMqPublishWorker()
	handler.startMqttPublishWorker()
	return handler
}


func (h *Handler) startMqPublishWorker() {
    h.wg.Add(1)
    go func() {
        defer h.wg.Done()
        for h.running {
            job, ok := <-h.mqJobQueue
			if !ok {
				return
			}
            err := h.mqPublisher.Publish(job.topic, []byte(job.key), job.payload)
            if err != nil {
                 log.Printf("Failed to publish to MQ: %v, topic=%s, key=%q", err, job.topic, job.key)
            } else {
                 log.Printf("Send Msg to MQ, topic=%s, key=%q, payload=%q", job.topic, job.key, job.payload)
			}
        }
    }()
}

func (h *Handler) startMqttPublishWorker() {
    h.wg.Add(1)
    go func() {
        defer h.wg.Done()
        for h.running {
            job, ok := <-h.mqttJobQueue
            if !ok {
                return
            }
            err := h.mqttPublisher.Publish(job.topic, job.payload)
            if err != nil {
                log.Printf("Failed to publish to MQTT: %v, topic=%s", err, job.topic)
            } else {
                log.Printf("Send Msg to MQTT, topic=%s, payload=%q", job.topic, job.payload)
			}
        }
    }()
}

func (h *Handler) Close() {
	h.running = false
	close(h.mqttJobQueue)
	close(h.mqJobQueue)
	h.wg.Wait()
}

// Handles telemetry data from edge (received via MQTT), and forwards to MQ for further processing.
func (h *Handler) EdgeGatewayTelemetryMsgHandler(_ mqtt.Client, msg mqtt.Message) {
    if h.mqPublisher == nil {
        log.Printf("MQ publisher not set")
        return
    }

    edgeGatewayID := strings.Split(msg.Topic(), "/")[1]
	if edgeGatewayID == "" {
        log.Printf("edgeGatewayID not found")
        return
	}

	raw_data := msg.Payload() // TODO: deserialize raw telemetry data here ^_^
    job := mqJob{
        topic:      "cloud.telemetry",
        key:        []byte(edgeGatewayID),
        payload:    raw_data,
    }

    select {
    case h.mqJobQueue <- job:
		{}
    default:
        log.Printf("MQ job queue is full, dropping telemetry msg for topic: %s", msg.Topic())
    }
}

// Handles simple data from edge (received via MQTT), and forwards to MQ for further processing.
func (h *Handler) MqttSimpleMsgHandler(msg mqtt.Message, dstTopic string) {
    if h.mqPublisher == nil {
        log.Printf("MQ publisher not set")
        return
    }

    edgeGatewayID := strings.Split(msg.Topic(), "/")[1]
	if edgeGatewayID == "" {
        log.Printf("edgeGatewayID not found")
        return
	}

    job := mqJob{
        topic:      dstTopic,
        key:        []byte(edgeGatewayID),
        payload:    msg.Payload(),
    }

    select {
    case h.mqJobQueue <- job:
		{}
    default:
        log.Printf("MQ job queue is full, dropping msg for topic: %s", msg.Topic())
    }
}

// Handles device model data from edge (received via MQTT), and forwards to MQ
func (h *Handler) EdgeGatewayModelMsgHandler(_ mqtt.Client, msg mqtt.Message) {
	h.MqttSimpleMsgHandler(msg, "cloud.model")
}

// Handles alert msg from edge (received via MQTT), and forwards to MQ
func (h *Handler) EdgeGatewayAlertMsgHandler(_ mqtt.Client, msg mqtt.Message) {
	h.MqttSimpleMsgHandler(msg, "cloud.alert")
}

// Handles any other common msg from edge (received via MQTT), and forwards to MQ
func (h *Handler) EdgeGatewayMsgHandler(_ mqtt.Client, msg mqtt.Message) {
	h.MqttSimpleMsgHandler(msg, "cloud.notify")
}


// Handles downstream message from cloud (received via MQ), and forwards to MQTT
func (h *Handler) CloudMsgHandler(topic string, key []byte, value []byte) {
    if h.mqttPublisher == nil {
        log.Printf("MQTT publisher not set")
        return
    }
	if (key == nil) || string(key) == "" {
        log.Printf("edgeGatewayID not found")
		return
	}
	edgeGatewayID := string(key)
	mqttTopic := fmt.Sprintf("egw/%s/notify", edgeGatewayID)
    job := mqttJob{
        topic:   mqttTopic,
        payload: value,
    }

    select {
    case h.mqttJobQueue <- job:
		{}
    default:
        log.Printf("MQTT job queue is full, dropping msg for topic: %s, key=%s", topic, string(key))
	}
}
