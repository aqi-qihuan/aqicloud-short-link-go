package model

import "time"

// TrafficTaskDO maps to the traffic_task table.
type TrafficTaskDO struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AccountNo   int64     `gorm:"column:account_no" json:"account_no"`
	TrafficID   int64     `gorm:"column:traffic_id" json:"traffic_id"`
	UseTimes    int       `gorm:"column:use_times" json:"use_times"`
	LockState   string    `gorm:"column:lock_state" json:"lock_state"`
	BizID       string    `gorm:"column:biz_id" json:"biz_id"`
	GmtCreate   time.Time `gorm:"column:gmt_create;autoCreateTime" json:"gmt_create"`
	GmtModified time.Time `gorm:"column:gmt_modified;autoUpdateTime" json:"gmt_modified"`
}

func (TrafficTaskDO) TableName() string {
	return "traffic_task"
}
