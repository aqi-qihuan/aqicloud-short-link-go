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

// StartAddMappingListener consumes from short_link.add.mapping.queue.
func StartAddMappingListener(rmq *mq.RabbitMQ, svc *service.ShortLinkService) {
	handler := func(body []byte) error {
		var eventMsg model.EventMessage
		if err := json.Unmarshal(body, &eventMsg); err != nil {
			log.Printf("[MQ] add_mapping unmarshal error: %v", err)
			return err
		}
		eventMsg.EventMessageType = string(enums.SHORT_LINK_ADD_MAPPING)
		log.Printf("[MQ] consuming add_mapping, messageId=%s", eventMsg.MessageId)
		svc.HandleAddShortLink(&eventMsg)
		return nil
	}
	if err := rmq.Consume(config.QueueAddMapping, "add_mapping_consumer", handler); err != nil {
		log.Fatalf("start add_mapping listener: %v", err)
	}
}
