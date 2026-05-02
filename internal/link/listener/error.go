package listener

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aqi/aqicloud-short-link-go/internal/common/alert"
	"github.com/aqi/aqicloud-short-link-go/internal/common/model"
	"github.com/aqi/aqicloud-short-link-go/internal/common/mq"
	"github.com/aqi/aqicloud-short-link-go/internal/link/config"
)

// StartErrorListener consumes from short_link.error.queue.
// Logs failed messages and sends alert notifications.
func StartErrorListener(rmq *mq.RabbitMQ, alerter alert.Alerter) {
	handler := func(body []byte) error {
		var eventMsg model.EventMessage
		if err := json.Unmarshal(body, &eventMsg); err != nil {
			log.Printf("[MQ-ERROR] unmarshal error message failed: %v, raw=%s", err, string(body))
			return nil // don't requeue unparseable messages
		}
		log.Printf("[MQ-ERROR] event=%s bizId=%s accountNo=%d messageId=%s content=%s",
			eventMsg.EventMessageType, eventMsg.BizId, eventMsg.AccountNo, eventMsg.MessageId, eventMsg.Content)

		// Send alert notification
		title := fmt.Sprintf("MQ Error: %s", eventMsg.EventMessageType)
		content := fmt.Sprintf("BizId: %s\nAccountNo: %d\nMessageId: %s\nContent: %s",
			eventMsg.BizId, eventMsg.AccountNo, eventMsg.MessageId, eventMsg.Content)
		if err := alerter.Send(title, content); err != nil {
			log.Printf("[ALERT] send alert failed: %v", err)
		}
		return nil
	}
	if err := rmq.Consume(config.QueueError, "error_consumer", handler); err != nil {
		log.Fatalf("start error listener: %v", err)
	}
}
