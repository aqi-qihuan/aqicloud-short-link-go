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

// StartUpdateMappingListener consumes from short_link.update.mapping.queue.
func StartUpdateMappingListener(rmq *mq.RabbitMQ, svc *service.ShortLinkService) {
	handler := func(body []byte) error {
		var eventMsg model.EventMessage
		if err := json.Unmarshal(body, &eventMsg); err != nil {
			log.Printf("[MQ] update_mapping unmarshal error: %v", err)
			return err
		}
		eventMsg.EventMessageType = string(enums.SHORT_LINK_UPDATE_MAPPING)
		log.Printf("[MQ] consuming update_mapping, messageId=%s", eventMsg.MessageId)
		svc.HandleUpdateShortLink(&eventMsg)
		return nil
	}
	if err := rmq.Consume(config.QueueUpdateMapping, "update_mapping_consumer", handler); err != nil {
		log.Fatalf("start update_mapping listener: %v", err)
	}
}
