package config

import (
	"log"

	"github.com/aqi/aqicloud-short-link-go/internal/common/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeTraffic = "traffic.event.exchange"
	ExchangeError   = "traffic.error.exchange"

	QueueFreeInit   = "traffic.free_init.queue"
	QueueReleaseDelay = "traffic.release.delay.queue"
	QueueRelease    = "traffic.release.queue"
	QueueOrderTraffic = "order.traffic.queue"
	QueueError      = "traffic.error.queue"

	RoutingKeyFreeInit     = "traffic.free_init.routing.key"
	RoutingKeyReleaseDelay = "traffic.release.delay.routing.key"
	RoutingKeyRelease      = "traffic.release.routing.key"
	RoutingKeyError        = "traffic.error.routing.key"
)

// SetupExchangesAndQueues declares all exchanges, queues, and bindings for the account service.
func SetupExchangesAndQueues(rmq *mq.RabbitMQ) {
	if err := rmq.DeclareExchange(ExchangeTraffic); err != nil {
		log.Fatalf("declare exchange %s: %v", ExchangeTraffic, err)
	}
	if err := rmq.DeclareExchange(ExchangeError); err != nil {
		log.Fatalf("declare exchange %s: %v", ExchangeError, err)
	}

	// Free init queue (user registration)
	declareAndBind(rmq, QueueFreeInit, RoutingKeyFreeInit, ExchangeTraffic)

	// Delay queue for traffic rollback (60s TTL -> dead-letter -> release queue)
	delayArgs := amqp.Table{
		"x-message-ttl":          int64(60000),
		"x-dead-letter-exchange": ExchangeTraffic,
		"x-dead-letter-routing-key": RoutingKeyRelease,
	}
	if _, err := rmq.DeclareQueue(QueueReleaseDelay, delayArgs); err != nil {
		log.Fatalf("declare queue %s: %v", QueueReleaseDelay, err)
	}
	if err := rmq.BindQueue(QueueReleaseDelay, RoutingKeyReleaseDelay, ExchangeTraffic); err != nil {
		log.Fatalf("bind queue %s: %v", QueueReleaseDelay, err)
	}

	// Release queue (dead-letter target, consumed by listener)
	declareAndBind(rmq, QueueRelease, RoutingKeyRelease, ExchangeTraffic)

	// Order traffic queue (from shop service)
	if _, err := rmq.DeclareQueue(QueueOrderTraffic, nil); err != nil {
		log.Fatalf("declare queue %s: %v", QueueOrderTraffic, err)
	}

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
