package mqttclient

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Client represents an MQTT client with auto-reconnect capability
type MqttClient struct {
	opts   *mqtt.ClientOptions
	client mqtt.Client
	topics map[string]mqtt.MessageHandler // Map of topics to their handlers
	mu     sync.RWMutex // Protects concurrent access to topics
	ctx    context.Context
	cancel context.CancelFunc
}

// Create a new MQTT client instance
func New() *MqttClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &MqttClient{
		topics: make(map[string]mqtt.MessageHandler),
		ctx:    ctx,
		cancel: cancel,
	}
}


func (c *MqttClient) GenerateClientID() string {
	id := uuid.Must(uuid.NewV7())
	return fmt.Sprintf("mqtt-bridge-%s", id)
}

// Init initializes topics and their handlers (not thread-safe, call before Connect)
func (c *MqttClient) Init(broker string, topicHandlers map[string]mqtt.MessageHandler) {
	c.opts = mqtt.NewClientOptions().AddBroker(broker)
	c.opts.SetClientID(c.GenerateClientID())
	c.opts.SetCleanSession(true)
	c.opts.SetAutoReconnect(true)

	for topic, handler := range topicHandlers {
		c.topics[topic] = handler
	}
}


func (c *MqttClient) Connect() error {
	c.client = mqtt.NewClient(c.opts)
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	log.Printf("Create MQTT connection done")

	c.Resubscribe()

	go c.monitorConnection()
	return nil
}


func (c *MqttClient) Disconnect() {
	c.cancel()
	c.client.Disconnect(250)
}

// Re-subscribe to all registered topics
func (c *MqttClient) Resubscribe() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for topic, handler := range c.topics {
		token := c.client.Subscribe(topic, 1, handler)
		if token.Wait() && token.Error() != nil {
			log.Printf("Subscribe topic %s failed: %v", topic, token.Error())
		} else {
			log.Printf("Subscribe topic %s OK", topic)
		}
	}
}

// Monitor connection status and handle reconnection regularly
func (c *MqttClient) monitorConnection() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if !c.client.IsConnected() {
				log.Println("Connection lost, reconnecting...")
				if token := c.client.Connect(); token.Error() != nil {
					log.Printf("Reconnection failed: %v", token.Error())
				} else {
					c.Resubscribe()
				}
			}
		}
	}
}
