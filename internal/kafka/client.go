package kafkaclient

import (
    "log"
    "sync"
    "time"

    "github.com/confluentinc/confluent-kafka-go/kafka"
)

// MessageHandler defines the function signature for handling consumed messages.
type MessageHandler func(topic string, key, value []byte)

// KafkaClient encapsulates a Confluent Kafka client.
type KafkaClient struct {
    pConfig         *kafka.ConfigMap
    cConfig         *kafka.ConfigMap
    producer        *kafka.Producer
    consumer        *kafka.Consumer
    topicHandlers   map[string]MessageHandler // Topic to handler mapping
    running         bool                      // Flag to control running state
    mu              sync.RWMutex              // Protects consumer and producer pointers
    reconnectInterval time.Duration           // Interval for reconnecting
}

// NewKafkaClient creates a new KafkaClient instance.
func NewKafkaClient(brokers string, groupId string, topicHandlers map[string]MessageHandler) *KafkaClient {
    return &KafkaClient{
        pConfig: &kafka.ConfigMap{
            "bootstrap.servers":          brokers,
            "acks":                       "all",
            "retries":                    3,
        },
        cConfig: &kafka.ConfigMap{
            "bootstrap.servers":          brokers,
            "group.id":                   groupId,
            "auto.offset.reset":          "latest",
            "enable.auto.commit":         true,
            "auto.commit.interval.ms":    1000,
			"socket.timeout.ms":          5000,
			"socket.connection.setup.timeout.ms":   5000,
        },
        topicHandlers:   topicHandlers,
        running:         true,
        reconnectInterval: 5 * time.Second,
    }
}

// Init initializes the producer and starts the consume loop.
func (k *KafkaClient) Init() error {
    producer, err := kafka.NewProducer(k.pConfig)
    if err != nil {
        return err
    }
    k.producer = producer

    // Start the auto-reconnecting consume loop.
    go k.startConsumeLoop()
    return nil
}

// startConsumeLoop runs a loop that automatically reconnects on failure.
func (k *KafkaClient) startConsumeLoop() {
    for k.isRunning() {
        err := k.connectAndConsume()
        if err != nil {
            log.Printf("Kafka connection/consume failed: %v. Retrying in %v...", err, k.reconnectInterval)
        }
        if k.isRunning() {
            time.Sleep(k.reconnectInterval)
        }
    }
}

// connectAndConsume creates a new consumer, subscribes to topics, and polls for messages.
func (k *KafkaClient) connectAndConsume() error {
    consumer, err := kafka.NewConsumer(k.cConfig)
    if err != nil {
        return err
    }

    // Set the new consumer.
    k.mu.Lock()
    k.consumer = consumer
    k.mu.Unlock()

    // Extract topic list from the handler map.
    topics := make([]string, 0, len(k.topicHandlers))
    for topic := range k.topicHandlers {
        topics = append(topics, topic)
    }

    // Subscribe to all topics.
    err = consumer.SubscribeTopics(topics, nil)
    if err != nil {
        consumer.Close()
        return err
    }

    // Main polling loop.
    run := true
    for run && k.isRunning() {
        ev := k.consumer.Poll(500)
        if ev == nil {
            continue
        }

        switch e := ev.(type) {
        case *kafka.Message:
            // Dispatch message to its registered handler.
            if handler, exists := k.topicHandlers[*e.TopicPartition.Topic]; exists && handler != nil {
                handler(*e.TopicPartition.Topic, e.Key, e.Value)
            } else {
                log.Printf("[Consumed] No handler for topic '%s': Key=%s, Value=%s", 
                    *e.TopicPartition.Topic, string(e.Key), string(e.Value))
            }

        case kafka.Error:
            log.Printf("Kafka error: %v", e)
            if e.IsFatal() {
                run = false // Exit to trigger reconnection
            }
        }
    }

    // Clean up the consumer.
    k.mu.Lock()
    if k.consumer == consumer {
        k.consumer.Close()
        k.consumer = nil
    }
    k.mu.Unlock()

    return nil
}

// Publish sends a message synchronously.
func (k *KafkaClient) Publish(topic string, key string, value []byte) error {
    k.mu.RLock()
    producer := k.producer
    k.mu.RUnlock()

    if producer == nil {
        return kafka.NewError(kafka.ErrQueueFull, "producer not ready", false)
    }

    deliveryChan := make(chan kafka.Event, 1)
    err := producer.Produce(&kafka.Message{
        TopicPartition: kafka.TopicPartition{
            Topic:     &topic,
            Partition: kafka.PartitionAny,
        },
        Key:   []byte(key),
        Value: value,
    }, deliveryChan)

    if err != nil {
        close(deliveryChan)
        return err
    }

    // Wait for delivery report.
    e := <-deliveryChan
    m := e.(*kafka.Message)
    close(deliveryChan)

    if m.TopicPartition.Error != nil {
        return m.TopicPartition.Error
    }
    return nil
}

// Close shuts down the producer and consumer gracefully.
func (k *KafkaClient) Close() {
    k.mu.Lock()
    k.running = false
    k.mu.Unlock()

    if k.producer != nil {
        k.producer.Close()
    }
    log.Println("Kafka client closed")
}

// isRunning checks if the client is still running.
func (k *KafkaClient) isRunning() bool {
    k.mu.RLock()
    defer k.mu.RUnlock()
    return k.running
}
