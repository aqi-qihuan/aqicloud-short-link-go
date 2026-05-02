package model

import "time"

// AccountDO maps to the account table.
type AccountDO struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AccountNo  int64     `gorm:"column:account_no;uniqueIndex" json:"account_no"`
	HeadImg    string    `gorm:"column:head_img" json:"head_img"`
	Phone      string    `gorm:"column:phone" json:"phone"`
	Pwd        string    `gorm:"column:pwd" json:"-"`
	Secret     string    `gorm:"column:secret" json:"-"`
	Mail       string    `gorm:"column:mail" json:"mail"`
	Username   string    `gorm:"column:username" json:"username"`
	Auth       string    `gorm:"column:auth" json:"auth"`
	GmtCreate  time.Time `gorm:"column:gmt_create;autoCreateTime" json:"gmt_create"`
	GmtModified time.Time `gorm:"column:gmt_modified;autoUpdateTime" json:"gmt_modified"`
}

func (AccountDO) TableName() string {
	return "account"
}
