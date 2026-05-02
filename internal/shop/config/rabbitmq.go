package config

import (
	"log"

	"github.com/aqi/aqicloud-short-link-go/internal/common/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeOrder = "order.event.exchange"
	ExchangeError = "order.error.exchange"

	QueueCloseDelay = "order.close.delay.queue"
	QueueClose      = "order.close.queue"
	QueueUpdate     = "order.update.queue"
	QueueTraffic    = "order.traffic.queue"
	QueueError      = "order.error.queue"

	RoutingKeyCloseDelay = "order.close.delay.routing.key"
	RoutingKeyClose      = "order.close.delay.key"
	RoutingKeyUpdate     = "order.update.traffic.routing.key"
	RoutingKeyError      = "order.error.routing.key"

	BindingClose  = "order.close.delay.key"
	BindingUpdate = "order.update.*.routing.key"
	BindingTraffic = "order.*.traffic.routing.key"
)

// SetupExchangesAndQueues declares all exchanges, queues, and bindings for the shop service.
func SetupExchangesAndQueues(rmq *mq.RabbitMQ) {
	if err := rmq.DeclareExchange(ExchangeOrder); err != nil {
		log.Fatalf("declare exchange %s: %v", ExchangeOrder, err)
	}
	if err := rmq.DeclareExchange(ExchangeError); err != nil {
		log.Fatalf("declare exchange %s: %v", ExchangeError, err)
	}

	// Delay queue: 60s TTL -> dead-letter -> close queue
	delayArgs := amqp.Table{
		"x-message-ttl":           int64(60000),
		"x-dead-letter-exchange":  ExchangeOrder,
		"x-dead-letter-routing-key": RoutingKeyClose,
	}
	if _, err := rmq.DeclareQueue(QueueCloseDelay, delayArgs); err != nil {
		log.Fatalf("declare queue %s: %v", QueueCloseDelay, err)
	}
	if err := rmq.BindQueue(QueueCloseDelay, RoutingKeyCloseDelay, ExchangeOrder); err != nil {
		log.Fatalf("bind queue %s: %v", QueueCloseDelay, err)
	}

	// Close queue (dead-letter target)
	declareAndBind(rmq, QueueClose, BindingClose, ExchangeOrder)

	// Update queue (order state update)
	declareAndBind(rmq, QueueUpdate, BindingUpdate, ExchangeOrder)

	// Traffic queue (traffic pack delivery)
	declareAndBind(rmq, QueueTraffic, BindingTraffic, ExchangeOrder)

	// Error queue
	declareAndBind(rmq, QueueError, RoutingKeyError, ExchangeError)
}

func declareAndBind(rmq *mq.RabbitMQ, queue, bindingKey, exchange string) {
	if _, err := rmq.DeclareQueue(queue, nil); err != nil {
		log.Fatalf("declare queue %s: %v", queue, err)
	}
	if err := rmq.BindQueue(queue, bindingKey, exchange); err != nil {
		log.Fatalf("bind queue %s to %s: %v", queue, exchange, err)
	}
}
