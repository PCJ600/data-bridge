package kafkaclient

import (
    "log"
    "github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaClient struct {
    pConfig  *kafka.ConfigMap
    cConfig  *kafka.ConfigMap
    producer *kafka.Producer
    consumer *kafka.Consumer
    topics   []string // Topics to subscribe
    handler  func(topic string, key, value []byte) // Callback for handling consumed messages
    running  bool // Flag to control running state
}


func NewKafkaClient(brokers string, groupId string, topics []string) *KafkaClient {
    return &KafkaClient{
        pConfig: &kafka.ConfigMap{
            "bootstrap.servers":          brokers,
            "acks":                       "all",
            "retries":                    3,
            "enable.idempotence":         true, // Enable idempotent producer to avoid duplicates
        },
        cConfig: &kafka.ConfigMap{
            "bootstrap.servers":          brokers,
            "group.id":                   groupId,
            "auto.offset.reset":          "latest", // Recommended for production
            "enable.auto.commit":         true,
            "auto.commit.interval.ms":    1000,

            // Uncomment below for SASL/SSL
            // "security.protocol":          "SASL_SSL",
            // "sasl.mechanisms":            "PLAIN",
            // "sasl.username":              "your-username",
            // "sasl.password":              "your-password",
        },
        topics:  topics,
        running: true,
    }
}

// Init initializes the producer and consumer instances.
func (k *KafkaClient) Init() error {
    producer, err := kafka.NewProducer(k.pConfig)
    if err != nil {
        return err
    }
    k.producer = producer

    consumer, err := kafka.NewConsumer(k.cConfig)
    if err != nil {
        return err
    }
    k.consumer = consumer

    return nil
}

// Connect starts consuming messages from the subscribed topics.
// It runs in a separate goroutine and handles message polling.
func (k *KafkaClient) Connect() error {
    err := k.consumer.SubscribeTopics(k.topics, nil)
    if err != nil {
        return err
    }

    go k.consumeLoop()
    return nil
}

// consumeLoop continuously polls for messages and handles them.
// It stops when k.running becomes false.
func (k *KafkaClient) consumeLoop() {
    for k.running {
        ev := k.consumer.Poll(500) // Wait up to 500ms
        if ev == nil {
            continue
        }

        switch e := ev.(type) {
        case *kafka.Message:
            if k.handler != nil {
                k.handler(*e.TopicPartition.Topic, e.Key, e.Value)
            } else {
                log.Printf("[Consumed] Topic: %s, Key: %s, Value: %s",
                    *e.TopicPartition.Topic, string(e.Key), string(e.Value))
            }

        case kafka.Error:
            if e.IsFatal() {
                log.Printf("Fatal consumer error: %v", e)
            } else {
                log.Printf("Consumer error: %v", e)
            }
        }
    }
}

// Publish sends a message synchronously (waits for delivery report).
// topic: target topic
// key: message key (used for partitioning)
// value: message payload
func (k *KafkaClient) Publish(topic, key string, value []byte) error {
    deliveryChan := make(chan kafka.Event, 1)
    err := k.producer.Produce(&kafka.Message{
        TopicPartition: kafka.TopicPartition{
            Topic:     &topic,
            Partition: kafka.PartitionAny,
        },
        Key:   []byte(key),
        Value: value,
    }, deliveryChan)

    if err != nil {
        return err
    }

    // Wait for delivery report
    e := <-deliveryChan
    m := e.(*kafka.Message)
    if m.TopicPartition.Error != nil {
        return m.TopicPartition.Error
    }
    close(deliveryChan)
    return nil
}

// Close shuts down the producer and consumer gracefully.
func (k *KafkaClient) Close() {
    k.running = false
    if k.producer != nil {
        k.producer.Close()
    }
    if k.consumer != nil {
        k.consumer.Close()
    }
    log.Println("Kafka client closed")
}
