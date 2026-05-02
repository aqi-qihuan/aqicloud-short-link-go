package request

// ConfirmOrderRequest is the body for POST /api/order/v1/confirm.
type ConfirmOrderRequest struct {
	ProductID         int64   `json:"productId"`
	BuyNum            int     `json:"buyNum"`
	ClientType        string  `json:"clientType"` // APP/PC/H5
	PayType           string  `json:"payType"`    // WECHAT_PAY/ALI_PAY
	TotalAmount       float64 `json:"totalAmount"`
	PayAmount         float64 `json:"payAmount"`
	Token             string  `json:"token"`             // anti-resubmit token
	BillType          string  `json:"billType"`
	BillHeader        string  `json:"billHeader"`
	BillContent       string  `json:"billContent"`
	BillReceiverPhone string  `json:"billReceiverPhone"`
	BillReceiverEmail string  `json:"billReceiverEmail"`
}

// ProductOrderPageRequest is the body for POST /api/order/v1/page.
type ProductOrderPageRequest struct {
	State string `json:"state"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
}
