package model

import (
	"time"
)

// ProductDO maps to the product table.
type ProductDO struct {
	ID          int64     `gorm:"column:id;primaryKey" json:"id"`
	Title       string    `gorm:"column:title" json:"title"`
	Detail      string    `gorm:"column:detail" json:"detail"`
	Img         string    `gorm:"column:img" json:"img"`
	Level       string    `gorm:"column:level" json:"level"`
	OldAmount   float64   `gorm:"column:old_amount" json:"oldAmount"`
	Amount      float64   `gorm:"column:amount" json:"amount"`
	PluginType  string    `gorm:"column:plugin_type" json:"pluginType"`
	DayTimes    int       `gorm:"column:day_times" json:"dayTimes"`
	TotalTimes  int       `gorm:"column:total_times" json:"totalTimes"`
	ValidDay    int       `gorm:"column:valid_day" json:"validDay"`
	GmtModified time.Time `gorm:"column:gmt_modified;autoUpdateTime" json:"gmtModified"`
	GmtCreate   time.Time `gorm:"column:gmt_create;autoCreateTime" json:"gmtCreate"`
}

func (ProductDO) TableName() string {
	return "product"
}
