package listener

import (
	"encoding/json"
	"log"

	"github.com/aqi/aqicloud-short-link-go/internal/common/model"
	"github.com/aqi/aqicloud-short-link-go/internal/common/mq"
	"github.com/aqi/aqicloud-short-link-go/internal/shop/config"
	"github.com/aqi/aqicloud-short-link-go/internal/shop/service"
)

// StartOrderListeners starts all order-related MQ consumers.
func StartOrderListeners(rmq *mq.RabbitMQ, svc *service.OrderService) {
	// Close queue consumer (delayed order close after 60s)
	startListener(rmq, config.QueueClose, "close_consumer", svc)

	// Update queue consumer (order state update + traffic delivery)
	startListener(rmq, config.QueueUpdate, "update_consumer", svc)

	// Error queue consumer
	startErrorListener(rmq)
}

func startListener(rmq *mq.RabbitMQ, queue, consumer string, svc *service.OrderService) {
	handler := func(body []byte) error {
		var eventMsg model.EventMessage
		if err := json.Unmarshal(body, &eventMsg); err != nil {
			log.Printf("[MQ] order listener unmarshal error on %s: %v", queue, err)
			return err
		}
		log.Printf("[MQ] consuming %s, messageId=%s, type=%s", queue, eventMsg.MessageId, eventMsg.EventMessageType)
		svc.HandleProductOrderMessage(&eventMsg)
		return nil
	}
	if err := rmq.Consume(queue, consumer, handler); err != nil {
		log.Fatalf("start order listener %s: %v", queue, err)
	}
}

func startErrorListener(rmq *mq.RabbitMQ) {
	handler := func(body []byte) error {
		var eventMsg model.EventMessage
		if err := json.Unmarshal(body, &eventMsg); err != nil {
			log.Printf("[MQ-ERROR] unmarshal failed: %v, raw=%s", err, string(body))
			return nil
		}
		log.Printf("[MQ-ERROR] event=%s bizId=%s accountNo=%d",
			eventMsg.EventMessageType, eventMsg.BizId, eventMsg.AccountNo)
		return nil
	}
	if err := rmq.Consume(config.QueueError, "order_error_consumer", handler); err != nil {
		log.Fatalf("start order error listener: %v", err)
	}
}
