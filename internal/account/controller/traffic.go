package controller

import (
	"strconv"

	"github.com/aqi/aqicloud-short-link-go/internal/account/request"
	"github.com/aqi/aqicloud-short-link-go/internal/account/service"
	"github.com/aqi/aqicloud-short-link-go/internal/account/vo"
	"github.com/aqi/aqicloud-short-link-go/internal/common/enums"
	"github.com/aqi/aqicloud-short-link-go/internal/common/interceptor"
	"github.com/aqi/aqicloud-short-link-go/internal/common/response"
	"github.com/gin-gonic/gin"
)

type TrafficController struct {
	svc *service.TrafficService
}

func NewTrafficController(svc *service.TrafficService) *TrafficController {
	return &TrafficController{svc: svc}
}

// Reduce handles POST /api/traffic/v1/reduce (RPC internal).
func (ctrl *TrafficController) Reduce(c *gin.Context) {
	var req request.UseTrafficRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("invalid request body"))
		return
	}
	if req.AccountNo == 0 || req.BizID == "" {
		response.JSON(c, response.BuildError("accountNo and bizId are required"))
		return
	}

	if err := ctrl.svc.Reduce(&req); err != nil {
		response.JSON(c, response.BuildResult(enums.TRAFFIC_REDUCE_FAIL))
		return
	}

	response.JSON(c, response.BuildSuccess())
}

// Page handles GET /api/traffic/v1/page.
func (ctrl *TrafficController) Page(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	packs, total, err := ctrl.svc.Page(loginUser.AccountNo, page, size)
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}

	// Convert to VO
	vos := make([]vo.TrafficVO, len(packs))
	for i, p := range packs {
		vos[i] = vo.TrafficVO{
			ID:          p.ID,
			DayLimit:    p.DayLimit,
			DayUsed:     p.DayUsed,
			TotalLimit:  p.TotalLimit,
			AccountNo:   p.AccountNo,
			OutTradeNo:  p.OutTradeNo,
			Level:       p.Level,
			ExpiredDate: p.ExpiredDate,
			PluginType:  p.PluginType,
			ProductID:   p.ProductID,
			GmtCreate:   p.GmtCreate,
		}
	}

	response.JSON(c, response.BuildSuccessData(gin.H{
		"page":  page,
		"size":  size,
		"total": total,
		"list":  vos,
	}))
}

// Detail handles GET /api/traffic/v1/detail/:trafficId.
func (ctrl *TrafficController) Detail(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}

	trafficId, err := strconv.ParseInt(c.Param("trafficId"), 10, 64)
	if err != nil {
		response.JSON(c, response.BuildError("invalid trafficId"))
		return
	}

	pack, err := ctrl.svc.Detail(loginUser.AccountNo, trafficId)
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}

	vo := vo.TrafficVO{
		ID:          pack.ID,
		DayLimit:    pack.DayLimit,
		DayUsed:     pack.DayUsed,
		TotalLimit:  pack.TotalLimit,
		AccountNo:   pack.AccountNo,
		OutTradeNo:  pack.OutTradeNo,
		Level:       pack.Level,
		ExpiredDate: pack.ExpiredDate,
		PluginType:  pack.PluginType,
		ProductID:   pack.ProductID,
		GmtCreate:   pack.GmtCreate,
	}

	response.JSON(c, response.BuildSuccessData(vo))
}
