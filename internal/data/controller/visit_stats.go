package controller

import (
	"github.com/aqi/aqicloud-short-link-go/internal/common/enums"
	"github.com/aqi/aqicloud-short-link-go/internal/common/interceptor"
	"github.com/aqi/aqicloud-short-link-go/internal/common/response"
	"github.com/aqi/aqicloud-short-link-go/internal/data/request"
	"github.com/aqi/aqicloud-short-link-go/internal/data/service"
	"github.com/gin-gonic/gin"
)

type VisitStatsController struct {
	svc *service.VisitStatsService
}

func NewVisitStatsController(svc *service.VisitStatsService) *VisitStatsController {
	return &VisitStatsController{svc: svc}
}

// PageRecord handles POST /api/visit_stats/v1/page_record.
func (ctrl *VisitStatsController) PageRecord(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	var req request.VisitRecordPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("invalid request body"))
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}

	result, err := ctrl.svc.PageRecord(loginUser.AccountNo, req.Code, req.Page, req.Size)
	if err != nil {
		response.JSON(c, response.BuildResult(enums.DATA_OUT_OF_LIMIT_SIZE))
		return
	}
	response.JSON(c, response.BuildSuccessData(result))
}

// RegionDay handles POST /api/visit_stats/v1/region_day.
func (ctrl *VisitStatsController) RegionDay(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	var req request.RegionQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("invalid request body"))
		return
	}

	result, err := ctrl.svc.RegionDay(loginUser.AccountNo, req.Code, req.StartTime, req.EndTime)
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}
	response.JSON(c, response.BuildSuccessData(result))
}

// Trend handles POST /api/visit_stats/v1/trend.
func (ctrl *VisitStatsController) Trend(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	var req request.VisitTrendQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("invalid request body"))
		return
	}

	result, err := ctrl.svc.Trend(loginUser.AccountNo, req.Code, req.Type, req.StartTime, req.EndTime)
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}
	response.JSON(c, response.BuildSuccessData(result))
}

// FrequentIP handles POST /api/visit_stats/v1/frequent_ip.
func (ctrl *VisitStatsController) FrequentIP(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	var req request.FrequentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("invalid request body"))
		return
	}

	result, err := ctrl.svc.FrequentIP(loginUser.AccountNo, req.Code)
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}
	response.JSON(c, response.BuildSuccessData(result))
}

// FrequentReferer handles POST /api/visit_stats/v1/frequent_referer.
func (ctrl *VisitStatsController) FrequentReferer(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	var req request.FrequentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("invalid request body"))
		return
	}

	result, err := ctrl.svc.FrequentReferer(loginUser.AccountNo, req.Code)
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}
	response.JSON(c, response.BuildSuccessData(result))
}

// DeviceInfo handles POST /api/visit_stats/v1/device_info.
func (ctrl *VisitStatsController) DeviceInfo(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	var req request.DeviceInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, response.BuildError("invalid request body"))
		return
	}

	result, err := ctrl.svc.DeviceInfo(loginUser.AccountNo, req.Code, req.Field)
	if err != nil {
		response.JSON(c, response.BuildError(err.Error()))
		return
	}
	response.JSON(c, response.BuildSuccessData(result))
}
