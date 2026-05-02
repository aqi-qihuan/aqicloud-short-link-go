package request

// AccountRegisterRequest is the request body for POST /api/account/v1/register.
type AccountRegisterRequest struct {
	HeadImg  string `json:"head_img"`
	Phone    string `json:"phone"`
	Pwd      string `json:"pwd"`
	Mail     string `json:"mail"`
	Username string `json:"username"`
	Code     string `json:"code"`
}

// AccountLoginRequest is the request body for POST /api/account/v1/login.
type AccountLoginRequest struct {
	Phone string `json:"phone"`
	Pwd   string `json:"pwd"`
}

// SendCodeRequest is the request body for POST /api/account/v1/send_code.
type SendCodeRequest struct {
	Captcha string `json:"captcha"`
	To      string `json:"to"`
}

// UseTrafficRequest is the request body for POST /api/traffic/v1/reduce.
type UseTrafficRequest struct {
	AccountNo int64  `json:"accountNo"`
	BizID     string `json:"bizId"`
}

// TrafficPageRequest is the query params for GET /api/traffic/v1/page.
type TrafficPageRequest struct {
	Page int `json:"page" form:"page"`
	Size int `json:"size" form:"size"`
}
