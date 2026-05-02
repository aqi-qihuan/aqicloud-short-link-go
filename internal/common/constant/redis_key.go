package constant

import "fmt"

const (
	// CHECK_CODE_KEY is the Redis key pattern for verification codes.
	// Format: code:SEND_CODE_ENUM:phone
	CHECK_CODE_KEY = "code:%s:%s"

	// SUBMIT_ORDER_TOKEN_KEY is the Redis key for order submission idempotency.
	// Format: order:submit:accountNo:requestToken
	SUBMIT_ORDER_TOKEN_KEY = "order:submit:%d:%s"

	// DAY_TOTAL_TRAFFIC is the Redis key for daily traffic quota counter.
	// Format: lock:traffic:day_total:accountNo
	DAY_TOTAL_TRAFFIC = "lock:traffic:day_total:%d"
)

// FormatCheckCodeKey formats the verification code Redis key.
func FormatCheckCodeKey(sendCodeType string, phone string) string {
	return fmt.Sprintf(CHECK_CODE_KEY, sendCodeType, phone)
}

// FormatSubmitOrderTokenKey formats the order token Redis key.
func FormatSubmitOrderTokenKey(accountNo int64, requestToken string) string {
	return fmt.Sprintf(SUBMIT_ORDER_TOKEN_KEY, accountNo, requestToken)
}

// FormatDayTotalTrafficKey formats the daily traffic Redis key.
func FormatDayTotalTrafficKey(accountNo int64) string {
	return fmt.Sprintf(DAY_TOTAL_TRAFFIC, accountNo)
}
