package vo

import "time"

// AccountVO is the response for account detail.
type AccountVO struct {
	AccountNo int64     `json:"account_no"`
	HeadImg   string    `json:"head_img"`
	Phone     string    `json:"phone"`
	Mail      string    `json:"mail"`
	Username  string    `json:"username"`
	Auth      string    `json:"auth"`
	GmtCreate time.Time `json:"create_time"`
}

// TrafficVO is the response for traffic detail.
type TrafficVO struct {
	ID          int64      `json:"id"`
	DayLimit    int        `json:"day_limit"`
	DayUsed     int        `json:"day_used"`
	TotalLimit  int        `json:"total_limit"`
	AccountNo   int64      `json:"account_no"`
	OutTradeNo  string     `json:"out_trade_no"`
	Level       string     `json:"level"`
	ExpiredDate *time.Time `json:"expired_date"`
	PluginType  string     `json:"plugin_type"`
	ProductID   int64      `json:"product_id"`
	GmtCreate   time.Time  `json:"gmt_create"`
}
