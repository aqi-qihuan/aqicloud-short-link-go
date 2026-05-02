package mq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer wraps a Kafka writer for producing messages.
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer creates a new Kafka producer.
func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10,
		RequiredAcks: kafka.RequireOne,
	}
	return &KafkaProducer{writer: w}
}

// PublishJSON sends a JSON message to Kafka.
func (p *KafkaProducer) PublishJSON(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: data,
	})
	if err != nil {
		log.Printf("[Kafka] publish error: %v", err)
	}
	return err
}

// Close closes the Kafka writer.
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
