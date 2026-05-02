package listener

import (
	"encoding/json"
	"log"

	"github.com/aqi/aqicloud-short-link-go/internal/account/config"
	"github.com/aqi/aqicloud-short-link-go/internal/account/service"
	"github.com/aqi/aqicloud-short-link-go/internal/common/model"
	"github.com/aqi/aqicloud-short-link-go/internal/common/mq"
)

// StartTrafficListeners starts all traffic-related MQ consumers.
func StartTrafficListeners(rmq *mq.RabbitMQ, svc *service.TrafficService) {
	// Consumer for free init events (user registration)
	startListener(rmq, config.QueueFreeInit, "free_init_consumer", svc)

	// Consumer for order pay events (from shop service)
	startListener(rmq, config.QueueOrderTraffic, "order_traffic_consumer", svc)

	// Consumer for traffic release/rollback events (dead-letter from delay queue)
	startListener(rmq, config.QueueRelease, "release_consumer", svc)

	// Error queue consumer
	startErrorListener(rmq)
}

func startListener(rmq *mq.RabbitMQ, queue, consumer string, svc *service.TrafficService) {
	handler := func(body []byte) error {
		var eventMsg model.EventMessage
		if err := json.Unmarshal(body, &eventMsg); err != nil {
			log.Printf("[MQ] traffic listener unmarshal error on %s: %v", queue, err)
			return err
		}
		log.Printf("[MQ] consuming %s, messageId=%s, type=%s", queue, eventMsg.MessageId, eventMsg.EventMessageType)
		svc.HandleTrafficMessage(&eventMsg)
		return nil
	}
	if err := rmq.Consume(queue, consumer, handler); err != nil {
		log.Fatalf("start traffic listener %s: %v", queue, err)
	}
}

func startErrorListener(rmq *mq.RabbitMQ) {
	handler := func(body []byte) error {
		var eventMsg model.EventMessage
		if err := json.Unmarshal(body, &eventMsg); err != nil {
			log.Printf("[MQ-ERROR] unmarshal failed: %v, raw=%s", err, string(body))
			return nil
		}
		log.Printf("[MQ-ERROR] event=%s bizId=%s accountNo=%d messageId=%s",
			eventMsg.EventMessageType, eventMsg.BizId, eventMsg.AccountNo, eventMsg.MessageId)
		return nil
	}
	if err := rmq.Consume(config.QueueError, "traffic_error_consumer", handler); err != nil {
		log.Fatalf("start traffic error listener: %v", err)
	}
}
