package model

// LoginUser represents the authenticated user context.
// Stored in gin.Context via the login interceptor.
type LoginUser struct {
	AccountNo int64  `json:"account_no"`
	HeadImg   string `json:"head_img"`
	Username  string `json:"username"`
	Mail      string `json:"mail"`
	Phone     string `json:"phone"`
	Auth      string `json:"auth"`
}
