package controller

import (
	"net/http"
	"regexp"

	"github.com/aqi/aqicloud-short-link-go/internal/common/model"
	"github.com/aqi/aqicloud-short-link-go/internal/common/mq"
	"github.com/aqi/aqicloud-short-link-go/internal/common/util"
	"github.com/aqi/aqicloud-short-link-go/internal/link/sharding"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LinkApiController struct {
	dbs        []*gorm.DB // 3 sharded DBs: ds0, ds1, dsa
	kafka      *mq.KafkaProducer
	codeRegexp *regexp.Regexp
}

func NewLinkApiController(dbs []*gorm.DB, kafka *mq.KafkaProducer) *LinkApiController {
	return &LinkApiController{
		dbs:        dbs,
		kafka:      kafka,
		codeRegexp: regexp.MustCompile(`^[a-z0-9A-Z]+$`),
	}
}

// Dispatch handles GET /{shortLinkCode} -> 302 redirect.
// This is the hot path for short link resolution.
func (ctrl *LinkApiController) Dispatch(c *gin.Context) {
	code := c.Param("shortLinkCode")
	if code == "" || !ctrl.codeRegexp.MatchString(code) {
		c.String(http.StatusBadRequest, "invalid short link code")
		return
	}

	// Route to correct shard: first char -> DB, last char -> table
	dbPrefix, tableSuffix := sharding.RouteShortLink(code)
	dbIdx := sharding.GetDBIndexByPrefix(dbPrefix)
	tableName := sharding.GetTableName("short_link", tableSuffix)

	var shortLink struct {
		OriginalUrl string `gorm:"column:original_url"`
		Expired     *string `gorm:"column:expired"`
		State       string `gorm:"column:state"`
		Del         int    `gorm:"column:del"`
		Code        string `gorm:"column:code"`
	}
	err := ctrl.dbs[dbIdx].Table(tableName).
		Where("code = ? AND del = 0", code).
		First(&shortLink).Error
	if err != nil {
		c.String(http.StatusNotFound, "short link not found")
		return
	}

	// Send visit log to Kafka
	ip := c.ClientIP()
	logRecord := model.LogRecord{
		IP:    ip,
		Ts:    util.GetCurrentTimestamp(),
		Event: "shortLinkVisit",
		BizId: code,
		Data: map[string]interface{}{
			"code": code,
			"ip":   ip,
		},
	}
	if ctrl.kafka != nil {
		go ctrl.kafka.PublishJSON(c, code, logRecord)
	}

	// Check if link is visitable
	if shortLink.Del == 1 || shortLink.State == "LOCK" {
		c.String(http.StatusForbidden, "short link is locked or deleted")
		return
	}

	// Strip URL prefix (snowflakeId&url)
	originalUrl := util.RemoveUrlPrefix(shortLink.OriginalUrl)

	// 302 redirect (not 301, to enable click tracking)
	c.Redirect(http.StatusFound, originalUrl)
}
