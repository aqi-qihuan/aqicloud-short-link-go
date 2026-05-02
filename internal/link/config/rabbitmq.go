package config

import (
	"log"

	"github.com/aqi/aqicloud-short-link-go/internal/common/mq"
)

const (
	ExchangeShortLink = "short_link.event.exchange"
	ExchangeError     = "short_link.error.exchange"

	QueueAddLink      = "short_link.add.link.queue"
	QueueAddMapping   = "short_link.add.mapping.queue"
	QueueDelLink      = "short_link.del.link.queue"
	QueueDelMapping   = "short_link.del.mapping.queue"
	QueueUpdateLink   = "short_link.update.link.queue"
	QueueUpdateMapping = "short_link.update.mapping.queue"
	QueueError        = "short_link.error.queue"

	RoutingKeyAdd    = "short_link.add.link.mapping.routing.key"
	RoutingKeyDel    = "short_link.del.link.mapping.routing.key"
	RoutingKeyUpdate = "short_link.update.link.mapping.routing.key"
	RoutingKeyError  = "short_link.error.routing.key"

	BindingAddLink      = "short_link.add.link.*.routing.key"
	BindingAddMapping   = "short_link.add.*.mapping.routing.key"
	BindingDelLink      = "short_link.del.link.*.routing.key"
	BindingDelMapping   = "short_link.del.*.mapping.routing.key"
	BindingUpdateLink   = "short_link.update.link.*.routing.key"
	BindingUpdateMapping = "short_link.update.*.mapping.routing.key"
)

// SetupExchangesAndQueues declares all exchanges, queues, and bindings for the link service.
func SetupExchangesAndQueues(rmq *mq.RabbitMQ) {
	// Main exchange
	if err := rmq.DeclareExchange(ExchangeShortLink); err != nil {
		log.Fatalf("declare exchange %s: %v", ExchangeShortLink, err)
	}
	// Error exchange
	if err := rmq.DeclareExchange(ExchangeError); err != nil {
		log.Fatalf("declare exchange %s: %v", ExchangeError, err)
	}

	// Add queues
	declareAndBind(rmq, QueueAddLink, BindingAddLink, ExchangeShortLink)
	declareAndBind(rmq, QueueAddMapping, BindingAddMapping, ExchangeShortLink)

	// Delete queues
	declareAndBind(rmq, QueueDelLink, BindingDelLink, ExchangeShortLink)
	declareAndBind(rmq, QueueDelMapping, BindingDelMapping, ExchangeShortLink)

	// Update queues
	declareAndBind(rmq, QueueUpdateLink, BindingUpdateLink, ExchangeShortLink)
	declareAndBind(rmq, QueueUpdateMapping, BindingUpdateMapping, ExchangeShortLink)

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
