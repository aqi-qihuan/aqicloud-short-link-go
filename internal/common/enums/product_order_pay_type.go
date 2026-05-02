package enums

type ProductOrderPayType string

const (
	PAY_WECHAT ProductOrderPayType = "WECHAT_PAY"
	PAY_ALI    ProductOrderPayType = "ALI_PAY"
	PAY_JD     ProductOrderPayType = "JD_PAY"
)
