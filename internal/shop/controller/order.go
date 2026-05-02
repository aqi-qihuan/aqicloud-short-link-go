package controller

import (
	"github.com/aqi/aqicloud-short-link-go/internal/common/enums"
	"github.com/aqi/aqicloud-short-link-go/internal/common/interceptor"
	"github.com/aqi/aqicloud-short-link-go/internal/common/response"
	"github.com/aqi/aqicloud-short-link-go/internal/shop/request"
	"github.com/aqi/aqicloud-short-link-go/internal/shop/service"
	"github.com/gin-gonic/gin"
)

type OrderController struct {
	svc *service.OrderService
}

func NewOrderController(svc *service.OrderService) *OrderController {
	return &OrderController{svc: svc}
}

// GetToken handles GET /api/order/v1/get_token.
func (ctrl *OrderController) GetToken(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	token, err := ctrl.svc.GetOrderToken(loginUser.AccountNo)
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}
	response.JSON(c, response.BuildSuccessData(token))
}

// Page handles POST /api/order/v1/page.
func (ctrl *OrderController) Page(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	var req request.ProductOrderPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("invalid request body"))
		return
	}
	result, err := ctrl.svc.Page(loginUser.AccountNo, &req)
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}
	response.JSON(c, response.BuildSuccessData(result))
}

// QueryState handles GET /api/order/v1/query_state.
func (ctrl *OrderController) QueryState(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	outTradeNo := c.Query("out_trade_no")
	if outTradeNo == "" {
		response.JSON(c, response.BuildError("out_trade_no is required"))
		return
	}
	state, err := ctrl.svc.QueryState(loginUser.AccountNo, outTradeNo)
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}
	response.JSON(c, response.BuildSuccessData(state))
}

// Confirm handles POST /api/order/v1/confirm.
func (ctrl *OrderController) Confirm(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	var req request.ConfirmOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("invalid request body"))
		return
	}
	result, err := ctrl.svc.Confirm(loginUser.AccountNo, loginUser.Username, &req)
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}
	response.JSON(c, response.BuildSuccessData(result))
}
