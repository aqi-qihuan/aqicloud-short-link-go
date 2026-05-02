package enums

type ProductOrderState string

const (
	ORDER_NEW    ProductOrderState = "NEW"
	ORDER_PAY    ProductOrderState = "PAY"
	ORDER_CANCEL ProductOrderState = "CANCEL"
)
