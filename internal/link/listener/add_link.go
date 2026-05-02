package listener

import (
	"encoding/json"
	"log"

	"github.com/aqi/aqicloud-short-link-go/internal/common/enums"
	"github.com/aqi/aqicloud-short-link-go/internal/common/model"
	"github.com/aqi/aqicloud-short-link-go/internal/common/mq"
	"github.com/aqi/aqicloud-short-link-go/internal/link/config"
	"github.com/aqi/aqicloud-short-link-go/internal/link/service"
)

// StartAddLinkListener consumes from short_link.add.link.queue.
func StartAddLinkListener(rmq *mq.RabbitMQ, svc *service.ShortLinkService) {
	handler := func(body []byte) error {
		var eventMsg model.EventMessage
		if err := json.Unmarshal(body, &eventMsg); err != nil {
			log.Printf("[MQ] add_link unmarshal error: %v", err)
			return err
		}
		eventMsg.EventMessageType = string(enums.SHORT_LINK_ADD_LINK)
		log.Printf("[MQ] consuming add_link, messageId=%s", eventMsg.MessageId)
		svc.HandleAddShortLink(&eventMsg)
		return nil
	}
	if err := rmq.Consume(config.QueueAddLink, "add_link_consumer", handler); err != nil {
		log.Fatalf("start add_link listener: %v", err)
	}
}
