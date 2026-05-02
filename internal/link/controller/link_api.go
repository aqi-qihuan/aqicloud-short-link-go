package controller

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/aqi/aqicloud-short-link-go/internal/common/constant"
	"github.com/aqi/aqicloud-short-link-go/internal/common/model"
	"github.com/aqi/aqicloud-short-link-go/internal/common/mq"
	"github.com/aqi/aqicloud-short-link-go/internal/common/util"
	"github.com/aqi/aqicloud-short-link-go/internal/link/sharding"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const shortLinkCacheTTL = 10 * time.Minute

type LinkApiController struct {
	dbs        []*gorm.DB // 3 sharded DBs: ds0, ds1, dsa
	rdb        *redis.Client
	kafka      *mq.KafkaProducer
	codeRegexp *regexp.Regexp
}

func NewLinkApiController(dbs []*gorm.DB, kafka *mq.KafkaProducer, rdb *redis.Client) *LinkApiController {
	return &LinkApiController{
		dbs:        dbs,
		kafka:      kafka,
		rdb:        rdb,
		codeRegexp: regexp.MustCompile(`^[a-z0-9A-Z]+$`),
	}
}

// Dispatch handles GET /{shortLinkCode} -> 302 redirect.
// This is the hot path for short link resolution.
// Cache strategy: Cache-Aside with Redis Hash (10min TTL).
func (ctrl *LinkApiController) Dispatch(c *gin.Context) {
	code := c.Param("shortLinkCode")
	if code == "" || !ctrl.codeRegexp.MatchString(code) {
		c.String(http.StatusBadRequest, "invalid short link code")
		return
	}

	// Step 1: 尝试从 Redis 缓存获取
	cacheKey := constant.FormatShortLinkCacheKey(code)
	if ctrl.rdb != nil {
		cached, err := ctrl.rdb.HGetAll(c, cacheKey).Result()
		if err == nil && len(cached) > 0 {
			// 缓存命中
			if cached["del"] == "1" || cached["state"] == "LOCK" {
				c.String(http.StatusForbidden, "short link is locked or deleted")
				return
			}
			ctrl.sendVisitLog(c, code)
			originalUrl := util.RemoveUrlPrefix(cached["url"])
			c.Redirect(http.StatusFound, originalUrl)
			return
		}
	}

	// Step 2: 缓存未命中，回源 MySQL
	dbPrefix, tableSuffix := sharding.RouteShortLink(code)
	dbIdx := sharding.GetDBIndexByPrefix(dbPrefix)
	tableName := sharding.GetTableName("short_link", tableSuffix)

	var shortLink struct {
		OriginalUrl string  `gorm:"column:original_url"`
		Expired     *string `gorm:"column:expired"`
		State       string  `gorm:"column:state"`
		Del         int     `gorm:"column:del"`
		Code        string  `gorm:"column:code"`
	}
	err := ctrl.dbs[dbIdx].Table(tableName).
		Where("code = ? AND del = 0", code).
		First(&shortLink).Error
	if err != nil {
		// 缓存空结果 (防缓存穿透)，短 TTL
		if ctrl.rdb != nil {
			ctrl.rdb.HSet(c, cacheKey, "url", "", "del", "1", "state", "")
			ctrl.rdb.Expire(c, cacheKey, 1*time.Minute)
		}
		c.String(http.StatusNotFound, "short link not found")
		return
	}

	// Step 3: 写入 Redis 缓存
	if ctrl.rdb != nil {
		ctx := context.Background()
		ctrl.rdb.HSet(ctx, cacheKey, "url", shortLink.OriginalUrl, "del", shortLink.Del, "state", shortLink.State)
		ctrl.rdb.Expire(ctx, cacheKey, shortLinkCacheTTL)
	}

	ctrl.sendVisitLog(c, code)

	if shortLink.Del == 1 || shortLink.State == "LOCK" {
		c.String(http.StatusForbidden, "short link is locked or deleted")
		return
	}

	originalUrl := util.RemoveUrlPrefix(shortLink.OriginalUrl)
	c.Redirect(http.StatusFound, originalUrl)
}

// sendVisitLog sends visit log to Kafka asynchronously.
func (ctrl *LinkApiController) sendVisitLog(c *gin.Context, code string) {
	if ctrl.kafka == nil {
		return
	}
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
	go func() {
		if err := ctrl.kafka.PublishJSON(c, code, logRecord); err != nil {
			log.Printf("[Kafka] publish visit log error: %v", err)
		}
	}()
}
