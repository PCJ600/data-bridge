package mqttclient

import (
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
	mu     sync.RWMutex // Concurrent protection
	qos    byte
}


func New() *MqttClient {
	return &MqttClient{
		topics: make(map[string]mqtt.MessageHandler),
		qos: 1,
	}
}

// Initializes subscribe topics and handlers (not thread-safe, call before Connect)
func (c *MqttClient) Init(broker string, topicHandlers map[string]mqtt.MessageHandler) {
	c.opts = mqtt.NewClientOptions().AddBroker(broker)
	c.opts.SetClientID(fmt.Sprintf("mqtt-bridge-%s", uuid.Must(uuid.NewV7())))
	c.opts.SetKeepAlive(30 * time.Second)
	c.opts.SetCleanSession(true)
	c.opts.SetAutoReconnect(true)
	c.opts.SetOnConnectHandler(c.onConnect)
	c.opts.SetConnectionLostHandler(c.onConnectionLost)
	for topic, handler := range topicHandlers {
		c.topics[topic] = handler
	}

	if err := c.Connect(); err != nil {
		log.Printf("first MQTT connection failed: %v", err)
	}

	// Delay the autoReconnect go-routine to avoid repeat connection
	time.AfterFunc(5 * time.Second, func() {
		go c.AutoReconnect()
	})
}

func (c *MqttClient) onConnect(client mqtt.Client) {
	log.Printf("Create MQTT connection done")
	c.ResubscribeAllTopics()
}

func (c *MqttClient) onConnectionLost(client mqtt.Client, err error) {
	log.Printf("MQTT connection lost")
}

func (c *MqttClient) Connect() error {
	// Purge old connection first
	c.Disconnect()

	c.mu.Lock()
	defer c.mu.Unlock()
	c.client = mqtt.NewClient(c.opts)
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *MqttClient) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		c.client.Disconnect(250)
	}
}

func (c *MqttClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return false
	}
	return c.client.IsConnectionOpen()
}


// TODO: handle subscribe error
func (c *MqttClient) ResubscribeAllTopics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for topic, handler := range c.topics {
		token := c.client.Subscribe(topic, c.qos, handler)
		if token.Wait() && token.Error() != nil {
			log.Printf("Subscribe MQTT topic %s failed: %v", topic, token.Error())
		} else {
			log.Printf("Subscribe MQTT topic %s OK", topic)
		}
	}
}

// Monitor connection and handle reconnection regularly
func (c *MqttClient) AutoReconnect() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	retry_times := 0
	for {
		select {
		case <-ticker.C:
			if c.IsConnected() {
				retry_times = 0
				continue
			}

			log.Printf("Detect MQTT connection lost, reconnecting...")
			retry_times++
			if err := c.Connect(); err != nil {
				log.Printf("Reconnect MQTT failed: %v, retry_times: %d", err, retry_times)
			}
		}
	}
}
