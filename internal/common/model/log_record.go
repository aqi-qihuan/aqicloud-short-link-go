package model

// LogRecord is the Kafka visit log record, compatible with Java's LogRecord.
type LogRecord struct {
	IP      string                 `json:"ip"`
	Ts      int64                  `json:"ts"`
	Event   string                 `json:"event"`
	Udid    string                 `json:"udid"`
	BizId   string                 `json:"bizId"`
	Data    map[string]interface{} `json:"data"`
}
