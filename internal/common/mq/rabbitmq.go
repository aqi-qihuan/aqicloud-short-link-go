package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ wraps the AMQP connection and provides publish/consume helpers.
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQ connects to RabbitMQ and returns a wrapper.
func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial failed: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq channel failed: %w", err)
	}
	return &RabbitMQ{conn: conn, channel: ch}, nil
}

// Channel returns the underlying AMQP channel.
func (r *RabbitMQ) Channel() *amqp.Channel {
	return r.channel
}

// DeclareExchange declares a topic exchange.
func (r *RabbitMQ) DeclareExchange(name string) error {
	return r.channel.ExchangeDeclare(name, "topic", true, false, false, false, nil)
}

// DeclareQueue declares a durable queue.
func (r *RabbitMQ) DeclareQueue(name string, args amqp.Table) (amqp.Queue, error) {
	return r.channel.QueueDeclare(name, true, false, false, false, args)
}

// BindQueue binds a queue to an exchange with a routing key.
func (r *RabbitMQ) BindQueue(queueName, routingKey, exchangeName string) error {
	return r.channel.QueueBind(queueName, routingKey, exchangeName, false, nil)
}

// PublishJSON publishes a JSON-serializable message to an exchange.
func (r *RabbitMQ) PublishJSON(exchange, routingKey string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.channel.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Body:         data,
	})
}

// Consume starts consuming from a queue. The handler is called for each message.
func (r *RabbitMQ) Consume(queueName, consumerTag string, handler func([]byte) error) error {
	msgs, err := r.channel.Consume(queueName, consumerTag, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume failed: %w", err)
	}
	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("[RabbitMQ] consume error on %s: %v", queueName, err)
				msg.Nack(false, true) // requeue
			} else {
				msg.Ack(false)
			}
		}
	}()
	return nil
}

// Close closes the channel and connection.
func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
