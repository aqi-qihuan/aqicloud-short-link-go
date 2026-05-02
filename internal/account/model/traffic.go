package model

import "time"

// TrafficDO maps to the traffic table (sharded: traffic_0, traffic_1).
type TrafficDO struct {
	ID          int64      `gorm:"column:id;primaryKey" json:"id"`
	DayLimit    int        `gorm:"column:day_limit" json:"day_limit"`
	DayUsed     int        `gorm:"column:day_used" json:"day_used"`
	TotalLimit  int        `gorm:"column:total_limit" json:"total_limit"`
	AccountNo   int64      `gorm:"column:account_no" json:"account_no"`
	OutTradeNo  string     `gorm:"column:out_trade_no" json:"out_trade_no"`
	Level       string     `gorm:"column:level" json:"level"`
	ExpiredDate *time.Time `gorm:"column:expired_date" json:"expired_date"`
	PluginType  string     `gorm:"column:plugin_type" json:"plugin_type"`
	ProductID   int64      `gorm:"column:product_id" json:"product_id"`
	GmtCreate   time.Time  `gorm:"column:gmt_create;autoCreateTime" json:"gmt_create"`
	GmtModified time.Time  `gorm:"column:gmt_modified;autoUpdateTime" json:"gmt_modified"`
}

func (TrafficDO) TableName() string {
	return "traffic"
}
