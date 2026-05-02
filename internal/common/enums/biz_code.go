package enums

// BizCodeEnum represents business error codes, compatible with Java's BizCodeEnum.
type BizCodeEnum struct {
	code    int
	message string
}

func (e BizCodeEnum) Code() int      { return e.code }
func (e BizCodeEnum) Message() string { return e.message }

// Common
var (
	COMMON_REPEAT_OPERATION = BizCodeEnum{110001, "重复操作，请稍后再试"}
	COMMON_NETWORK_ERROR    = BizCodeEnum{110002, "网络异常"}
)

// Link Group
var (
	GROUP_REPEAT       = BizCodeEnum{230001, "组名已经存在"}
	GROUP_NOT_EXIST    = BizCodeEnum{230002, "组名不存在"}
	GROUP_ADD_FAIL     = BizCodeEnum{230003, "新增组失败"}
	GROUP_OPERATE_FAIL = BizCodeEnum{230004, "操作组失败"}
)

// Verification Code
var (
	CODE_GENERATE_ERROR = BizCodeEnum{240001, "验证码生成失败"}
	CODE_LIMITED        = BizCodeEnum{240002, "验证码发送过快，请稍后再试"}
	CODE_ERROR          = BizCodeEnum{240003, "验证码错误"}
	CAPTCHA_ERROR       = BizCodeEnum{240004, "图形验证码错误"}
)

// Account
var (
	ACCOUNT_REPEAT      = BizCodeEnum{250001, "账号已经存在"}
	ACCOUNT_NOT_EXIST   = BizCodeEnum{250002, "账号不存在"}
	ACCOUNT_PWD_ERROR   = BizCodeEnum{250003, "账号密码错误"}
	ACCOUNT_UNLOGIN     = BizCodeEnum{250004, "账号未登录"}
	ACCOUNT_TOKEN_ERROR = BizCodeEnum{250005, "token不合法"}
)

// Short Link
var (
	SHORT_LINK_NOT_EXIST = BizCodeEnum{260001, "短链不存在"}
)

// Order
var (
	ORDER_PRICE_FAIL       = BizCodeEnum{280001, "价格不合法"}
	ORDER_REPEAT_SUBMIT    = BizCodeEnum{280002, "请勿重复提交"}
	ORDER_TOKEN_MISSING    = BizCodeEnum{280003, "token不存在，请重新获取"}
	ORDER_NOT_EXIST        = BizCodeEnum{280004, "订单不存在"}
	ORDER_STATE_ERROR      = BizCodeEnum{280005, "订单状态异常"}
	ORDER_CANCEL_SUCCESS   = BizCodeEnum{280006, "订单取消成功"}
	ORDER_CONFIRM_SUCCESS  = BizCodeEnum{280007, "订单确认成功"}
)

// Payment
var (
	PAY_ORDER_FAIL          = BizCodeEnum{300001, "创建支付订单失败"}
	PAY_CALLBACK_SIGN_FAIL  = BizCodeEnum{300002, "回调签名验证失败"}
	PAY_ORDER_STATE_ERROR   = BizCodeEnum{300003, "订单状态异常"}
	PAY_ORDER_TIMEOUT_ERROR = BizCodeEnum{300004, "订单超时"}
)

// Data
var (
	DATA_OUT_OF_LIMIT_SIZE = BizCodeEnum{400001, "数据量超过限制"}
	DATA_OUT_OF_LIMIT_DATE = BizCodeEnum{400002, "日期范围超过限制"}
)

// Flow Control
var (
	CONTROL_FLOW  = BizCodeEnum{500101, "限流控制"}
	CONTROL_DEGRADE = BizCodeEnum{500201, "降级控制"}
	CONTROL_AUTH  = BizCodeEnum{500301, "认证控制"}
)

// Traffic
var (
	TRAFFIC_NOT_EXIST    = BizCodeEnum{600001, "流量包不存在"}
	TRAFFIC_REDUCE_FAIL  = BizCodeEnum{600002, "流量包扣减失败"}
	TRAFFIC_DATA_ERROR   = BizCodeEnum{600003, "数据异常"}
)

// File Upload
var (
	FILE_UPLOAD_ERROR = BizCodeEnum{700001, "文件上传失败"}
)

// DB Routing
var (
	DB_ROUTING_ERROR = BizCodeEnum{800001, "数据库路由错误"}
)

// MQ
var (
	MQ_CONSUME_ERROR = BizCodeEnum{900001, "消息消费异常"}
)
