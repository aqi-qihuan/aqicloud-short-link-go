package controller

import (
	"github.com/aqi/aqicloud-short-link-go/internal/common/enums"
	"github.com/aqi/aqicloud-short-link-go/internal/common/interceptor"
	"github.com/aqi/aqicloud-short-link-go/internal/common/response"
	"github.com/aqi/aqicloud-short-link-go/internal/link/model"
	"github.com/aqi/aqicloud-short-link-go/internal/link/vo"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DomainController struct {
	db *gorm.DB
}

func NewDomainController(db *gorm.DB) *DomainController {
	return &DomainController{db: db}
}

// List handles GET /api/domain/v1/list.
func (ctrl *DomainController) List(c *gin.Context) {
	loginUser := interceptor.GetLoginUser(c)
	if loginUser == nil {
		response.JSON(c, response.BuildResult(enums.ACCOUNT_UNLOGIN))
		return
	}
	var domains []model.DomainDO
	ctrl.db.Where("del = 0").Find(&domains)
	list := make([]vo.DomainVO, len(domains))
	for i, d := range domains {
		list[i] = vo.DomainVO{
			ID:         d.ID,
			AccountNo:  d.AccountNo,
			DomainType: d.DomainType,
			Value:      d.Value,
		}
	}
	response.JSON(c, response.BuildSuccessData(list))
}
