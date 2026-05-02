package model

import "time"

// LinkGroupDO maps to the link_group table.
type LinkGroupDO struct {
	ID          int64     `gorm:"column:id;primaryKey" json:"id"`
	Title       string    `gorm:"column:title" json:"title"`
	AccountNo   int64     `gorm:"column:account_no" json:"account_no"`
	GmtCreate   time.Time `gorm:"column:gmt_create;autoCreateTime" json:"gmt_create"`
	GmtModified time.Time `gorm:"column:gmt_modified;autoUpdateTime" json:"gmt_modified"`
}

func (LinkGroupDO) TableName() string {
	return "link_group"
}
