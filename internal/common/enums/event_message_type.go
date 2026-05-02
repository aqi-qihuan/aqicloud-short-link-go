package enums

type EventMessageType string

const (
	SHORT_LINK_ADD           EventMessageType = "SHORT_LINK_ADD"
	SHORT_LINK_ADD_LINK      EventMessageType = "SHORT_LINK_ADD_LINK"
	SHORT_LINK_ADD_MAPPING   EventMessageType = "SHORT_LINK_ADD_MAPPING"
	SHORT_LINK_DEL           EventMessageType = "SHORT_LINK_DEL"
	SHORT_LINK_DEL_LINK      EventMessageType = "SHORT_LINK_DEL_LINK"
	SHORT_LINK_DEL_MAPPING   EventMessageType = "SHORT_LINK_DEL_MAPPING"
	SHORT_LINK_UPDATE        EventMessageType = "SHORT_LINK_UPDATE"
	SHORT_LINK_UPDATE_LINK   EventMessageType = "SHORT_LINK_UPDATE_LINK"
	SHORT_LINK_UPDATE_MAPPING EventMessageType = "SHORT_LINK_UPDATE_MAPPING"
	PRODUCT_ORDER_NEW        EventMessageType = "PRODUCT_ORDER_NEW"
	PRODUCT_ORDER_PAY        EventMessageType = "PRODUCT_ORDER_PAY"
	TRAFFIC_FREE_INIT        EventMessageType = "TRAFFIC_FREE_INIT"
	TRAFFIC_USED             EventMessageType = "TRAFFIC_USED"
)
