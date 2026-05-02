package model

// EventMessage is the RabbitMQ message envelope, compatible with Java's EventMessage.
type EventMessage struct {
	MessageId       string `json:"messageId"`
	EventMessageType string `json:"eventMessageType"`
	BizId           string `json:"bizId"`
	AccountNo       int64  `json:"accountNo"`
	Content         string `json:"content"`
}
