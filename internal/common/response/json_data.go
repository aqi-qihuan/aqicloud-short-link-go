package response

import (
	"net/http"

	"github.com/aqi/aqicloud-short-link-go/internal/common/enums"
	"github.com/gin-gonic/gin"
)

// JsonData is the unified API response format, compatible with Java's JsonData.
type JsonData struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

// BuildSuccess returns a success response with no data.
func BuildSuccess() *JsonData {
	return &JsonData{Code: 0, Data: nil, Msg: ""}
}

// BuildSuccessData returns a success response with data.
func BuildSuccessData(data interface{}) *JsonData {
	return &JsonData{Code: 0, Data: data, Msg: ""}
}

// BuildError returns an error response with message.
func BuildError(msg string) *JsonData {
	return &JsonData{Code: -1, Data: nil, Msg: msg}
}

// BuildCodeAndMsg returns an error response with code and message.
func BuildCodeAndMsg(code int, msg string) *JsonData {
	return &JsonData{Code: code, Data: nil, Msg: msg}
}

// BuildResult returns a response from a BizCodeEnum.
func BuildResult(bizCode enums.BizCodeEnum) *JsonData {
	return &JsonData{Code: bizCode.Code(), Data: nil, Msg: bizCode.Message()}
}

// JSON sends a JSON response.
func JSON(c *gin.Context, data *JsonData) {
	c.JSON(http.StatusOK, data)
}
