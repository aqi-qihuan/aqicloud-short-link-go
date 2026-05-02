package model

import "time"

// GroupCodeMappingDO maps to group_code_mapping table (sharded: _0, _1).
type GroupCodeMappingDO struct {
	ID          int64      `gorm:"column:id;primaryKey" json:"id"`
	GroupID     int64      `gorm:"column:group_id" json:"group_id"`
	Title       string     `gorm:"column:title" json:"title"`
	OriginalUrl string     `gorm:"column:original_url" json:"original_url"`
	Domain      string     `gorm:"column:domain" json:"domain"`
	Code        string     `gorm:"column:code" json:"code"`
	Sign        string     `gorm:"column:sign" json:"sign"`
	Expired     *time.Time `gorm:"column:expired" json:"expired"`
	AccountNo   int64      `gorm:"column:account_no" json:"account_no"`
	Del         int        `gorm:"column:del" json:"del"`
	State       string     `gorm:"column:state" json:"state"`
	LinkType    string     `gorm:"column:link_type" json:"link_type"`
	GmtCreate   time.Time  `gorm:"column:gmt_create;autoCreateTime" json:"gmt_create"`
	GmtModified time.Time  `gorm:"column:gmt_modified;autoUpdateTime" json:"gmt_modified"`
}

func (GroupCodeMappingDO) TableName() string {
	return "group_code_mapping"
}
