package controller

import (
	"strconv"

	"github.com/aqi/aqicloud-short-link-go/internal/common/enums"
	"github.com/aqi/aqicloud-short-link-go/internal/common/interceptor"
	"github.com/aqi/aqicloud-short-link-go/internal/common/response"
	"github.com/aqi/aqicloud-short-link-go/internal/common/util"
	"github.com/aqi/aqicloud-short-link-go/internal/link/model"
	"github.com/aqi/aqicloud-short-link-go/internal/link/request"
	"github.com/aqi/aqicloud-short-link-go/internal/link/vo"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LinkGroupController struct {
	dbs []*gorm.DB // ds0, ds1
}

func NewLinkGroupController(dbs []*gorm.DB) *LinkGroupController {
	return &LinkGroupController{dbs: dbs}
}

func (ctrl *LinkGroupController) getDB(accountNo int64) *gorm.DB {
	idx := int(accountNo%2) & 0x7FFFFFFF
	if idx >= len(ctrl.dbs) {
		idx = 0
	}
	return ctrl.dbs[idx]
}

// Add handles POST /api/group/v1/add.
func (ctrl *LinkGroupController) Add(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	var req request.LinkGroupAddRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Title == "" {
		response.JSON(c, response.BuildError("title is required"))
		return
	}
	group := model.LinkGroupDO{
		ID:        int64(util.GenerateSnowflakeID()),
		Title:     req.Title,
		AccountNo: loginUser.AccountNo,
	}
	db := ctrl.getDB(loginUser.AccountNo)
	if err := db.Create(&group).Error; err != nil {
		response.JSON(c, response.BuildResult(enums.GROUP_ADD_FAIL))
		return
	}
	response.JSON(c, response.BuildSuccessData(gin.H{"id": group.ID}))
}

// Del handles DELETE /api/group/v1/del/:group_id.
func (ctrl *LinkGroupController) Del(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	groupId, _ := strconv.ParseInt(c.Param("group_id"), 10, 64)
	db := ctrl.getDB(loginUser.AccountNo)
	result := db.Where("id = ? AND account_no = ?", groupId, loginUser.AccountNo).Delete(&model.LinkGroupDO{})
	if result.RowsAffected == 0 {
		response.JSON(c, response.BuildResult(enums.GROUP_NOT_EXIST))
		return
	}
	response.JSON(c, response.BuildSuccess())
}

// Detail handles GET /api/group/v1/detail/:group_id.
func (ctrl *LinkGroupController) Detail(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	groupId, _ := strconv.ParseInt(c.Param("group_id"), 10, 64)
	db := ctrl.getDB(loginUser.AccountNo)
	var group model.LinkGroupDO
	err := db.Where("id = ? AND account_no = ?", groupId, loginUser.AccountNo).First(&group).Error
	if err != nil {
		response.JSON(c, response.BuildResult(enums.GROUP_NOT_EXIST))
		return
	}
	response.JSON(c, response.BuildSuccessData(vo.LinkGroupVO{
		ID:          group.ID,
		Title:       group.Title,
		AccountNo:   group.AccountNo,
		GmtCreate:   group.GmtCreate,
		GmtModified: group.GmtModified,
	}))
}

// List handles GET /api/group/v1/list.
func (ctrl *LinkGroupController) List(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	db := ctrl.getDB(loginUser.AccountNo)
	var groups []model.LinkGroupDO
	db.Where("account_no = ?", loginUser.AccountNo).Find(&groups)
	list := make([]vo.LinkGroupVO, len(groups))
	for i, g := range groups {
		list[i] = vo.LinkGroupVO{
			ID:          g.ID,
			Title:       g.Title,
			AccountNo:   g.AccountNo,
			GmtCreate:   g.GmtCreate,
			GmtModified: g.GmtModified,
		}
	}
	response.JSON(c, response.BuildSuccessData(list))
}

// Update handles PUT /api/group/v1/update.
func (ctrl *LinkGroupController) Update(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	var req request.LinkGroupUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == 0 {
		response.JSON(c, response.BuildError("id and title are required"))
		return
	}
	db := ctrl.getDB(loginUser.AccountNo)
	result := db.Model(&model.LinkGroupDO{}).
		Where("id = ? AND account_no = ?", req.ID, loginUser.AccountNo).
		Update("title", req.Title)
	if result.RowsAffected == 0 {
		response.JSON(c, response.BuildResult(enums.GROUP_NOT_EXIST))
		return
	}
	response.JSON(c, response.BuildSuccess())
}
