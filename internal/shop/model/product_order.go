package model

import "time"

// ProductOrderDO maps to the product_order table.
type ProductOrderDO struct {
	ID                int64     `gorm:"column:id;primaryKey" json:"id"`
	ProductID         int64     `gorm:"column:product_id" json:"productId"`
	ProductTitle      string    `gorm:"column:product_title" json:"productTitle"`
	ProductAmount     float64   `gorm:"column:product_amount" json:"productAmount"`
	ProductSnapshot   string    `gorm:"column:product_snapshot" json:"productSnapshot"`
	BuyNum            int       `gorm:"column:buy_num" json:"buyNum"`
	OutTradeNo        string    `gorm:"column:out_trade_no" json:"outTradeNo"`
	State             string    `gorm:"column:state" json:"state"`
	CreateTime        time.Time `gorm:"column:create_time" json:"createTime"`
	TotalAmount       float64   `gorm:"column:total_amount" json:"totalAmount"`
	PayAmount         float64   `gorm:"column:pay_amount" json:"payAmount"`
	PayType           string    `gorm:"column:pay_type" json:"payType"`
	Nickname          string    `gorm:"column:nickname" json:"nickname"`
	AccountNo         int64     `gorm:"column:account_no" json:"accountNo"`
	Del               int       `gorm:"column:del" json:"del"`
	GmtModified       time.Time `gorm:"column:gmt_modified;autoUpdateTime" json:"gmtModified"`
	GmtCreate         time.Time `gorm:"column:gmt_create;autoCreateTime" json:"gmtCreate"`
	BillType          string    `gorm:"column:bill_type" json:"billType"`
	BillHeader        string    `gorm:"column:bill_header" json:"billHeader"`
	BillContent       string    `gorm:"column:bill_content" json:"billContent"`
	BillReceiverPhone string    `gorm:"column:bill_receiver_phone" json:"billReceiverPhone"`
	BillReceiverEmail string    `gorm:"column:bill_receiver_email" json:"billReceiverEmail"`
}

func (ProductOrderDO) TableName() string {
	return "product_order"
}
