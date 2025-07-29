package mqttclient

import (
	"fmt"
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Publisher interface {
    Publish(topic string, payload []byte) error
}

type MqttClient struct {
	broker         string
	clientID       string
	opts           *mqtt.ClientOptions
	client         mqtt.Client
	subscriptions  map[string]mqtt.MessageHandler
	mu             sync.RWMutex
	qos            byte
	retained       bool
}

// Creates a new MQTT client.
func New(broker string, clientID string) *MqttClient {
	return &MqttClient{
		broker: broker,
		clientID: clientID,
		subscriptions: make(map[string]mqtt.MessageHandler),
		qos: 1,
		retained: false,
	}
}

// Init initializes the MQTT client and connection.
func (c *MqttClient) Start() {
	c.opts = mqtt.NewClientOptions().AddBroker(c.broker)
	c.opts.SetClientID(c.clientID)
	c.opts.SetCleanSession(true)
	c.opts.SetAutoReconnect(true)
	c.opts.SetConnectRetryInterval(5 * time.Second)
	c.opts.SetOnConnectHandler(c.onConnect)
	c.opts.SetConnectionLostHandler(c.onConnectionLost)
	c.opts.SetKeepAlive(30 * time.Second)
	c.opts.SetWriteTimeout(5 * time.Second)

	if err := c.Connect(); err != nil {
		log.Printf("First MQTT connection failed: %v", err)
	}
}

func (c *MqttClient) onConnect(client mqtt.Client) {
	log.Printf("Create MQTT connection done")
	c.ResubscribeAllTopics()
}

func (c *MqttClient) onConnectionLost(client mqtt.Client, err error) {
	log.Printf("Detect MQTT connection lost")
}

func (c *MqttClient) Connect() error {
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

// Register subscription, execute all registered subscriptions after connecting to Broker.
func (c *MqttClient) RegisterSubscription(topic string, handler mqtt.MessageHandler) {
	c.subscriptions[topic] = handler
}

// TODO: handle subscribe error (or liveness probe)
func (c *MqttClient) ResubscribeAllTopics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for topic, handler := range c.subscriptions {
		token := c.client.Subscribe(topic, c.qos, handler)
		if token.Wait() && token.Error() != nil {
			log.Printf("Subscribe MQTT topic %s failed: %v", topic, token.Error())
		} else {
			log.Printf("Subscribe MQTT topic %s OK", topic)
		}
	}
}

func (c *MqttClient) Publish(topic string, payload []byte) error {
    if !c.IsConnected() {
        return fmt.Errorf("MQTT not connected")
    }

    token := c.client.Publish(topic, c.qos, c.retained, payload)
    if token.Wait() && token.Error() != nil {
        return token.Error()
    }
    return nil
}

