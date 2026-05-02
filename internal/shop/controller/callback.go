package controller

import (
	"context"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aqi/aqicloud-short-link-go/internal/common/enums"
	"github.com/aqi/aqicloud-short-link-go/internal/common/model"
	"github.com/aqi/aqicloud-short-link-go/internal/common/mq"
	"github.com/aqi/aqicloud-short-link-go/internal/common/util"
	shopdb "github.com/aqi/aqicloud-short-link-go/internal/shop/model"
	"github.com/aqi/aqicloud-short-link-go/internal/shop/component"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CallbackController struct {
	rdb     *redis.Client
	rmq     *mq.RabbitMQ
	payCfg  *component.PayConfig
	db      *gorm.DB // for direct order update when MQ unavailable
}

func NewCallbackController(rdb *redis.Client, rmq *mq.RabbitMQ, payCfg *component.PayConfig, db *gorm.DB) *CallbackController {
	return &CallbackController{rdb: rdb, rmq: rmq, payCfg: payCfg, db: db}
}

// WechatCallback handles POST /api/callback/order/v1/wechat.
func (ctrl *CallbackController) WechatCallback(c *gin.Context) {
	log.Printf("[WeChat Callback] received callback")

	// Parse XML body (WeChat Pay V2 uses XML)
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "FAIL", "message": "read body failed"})
		return
	}

	var notif struct {
		ReturnCode string `xml:"return_code"`
		ResultCode string `xml:"result_code"`
		OutTradeNo string `xml:"out_trade_no"`
		Sign       string `xml:"sign"`
	}
	if err := xml.Unmarshal(rawBody, &notif); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "FAIL", "message": "parse xml failed"})
		return
	}

	if notif.ReturnCode != "SUCCESS" {
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "忽略"})
		return
	}

	// Verify signature if WeChat API key is configured
	if ctrl.payCfg.WechatAPIKey != "" {
		params := xmlToMap(string(rawBody))
		if !component.VerifyWechatV2Sign(params, ctrl.payCfg.WechatAPIKey) {
			log.Printf("[WeChat Callback] signature verification failed")
			c.JSON(http.StatusOK, gin.H{"code": "FAIL", "message": "签名验证失败"})
			return
		}
		log.Printf("[WeChat Callback] signature verified")
	}

	outTradeNo := notif.OutTradeNo
	if outTradeNo != "" {
		ctrl.processOrderCallbackMsg(outTradeNo)
	}

	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}

// AlipayCallback handles POST /api/callback/order/v1/alipay.
func (ctrl *CallbackController) AlipayCallback(c *gin.Context) {
	log.Printf("[Alipay Callback] received callback")

	if err := c.Request.ParseForm(); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	// Collect all form params
	params := make(map[string]string)
	for k, v := range c.Request.Form {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	outTradeNo := params["out_trade_no"]
	tradeStatus := params["trade_status"]

	// Verify signature if Alipay public key is configured
	if ctrl.payCfg.AliPublicKey != "" {
		if !component.VerifyAlipaySign(params, ctrl.payCfg.AliPublicKey) {
			log.Printf("[Alipay Callback] signature verification failed for order: %s", outTradeNo)
			c.String(http.StatusOK, "fail")
			return
		}
		log.Printf("[Alipay Callback] signature verified for order: %s", outTradeNo)
	}

	if tradeStatus == "TRADE_SUCCESS" && outTradeNo != "" {
		ctrl.processOrderCallbackMsg(outTradeNo)
	}

	c.String(http.StatusOK, "success")
}

// processOrderCallbackMsg sends a PRODUCT_ORDER_PAY event to MQ.
func (ctrl *CallbackController) processOrderCallbackMsg(outTradeNo string) {
	// Idempotency guard via Redis setIfAbsent
	key := "pay:callback:" + outTradeNo
	set, err := ctrl.rdb.SetNX(context.Background(), key, "1", 0).Result()
	if err != nil || !set {
		log.Printf("[Callback] duplicate callback for order: %s", outTradeNo)
		return
	}

	// Try MQ first
	if ctrl.rmq != nil {
		eventMsg := model.EventMessage{
			MessageId:        util.GenerateUUID(),
			EventMessageType: string(enums.PRODUCT_ORDER_PAY),
			BizId:            outTradeNo,
		}
		if err := ctrl.rmq.PublishJSON("order.event.exchange", "order.update.traffic.routing.key", eventMsg); err != nil {
			log.Printf("[MQ] publish PRODUCT_ORDER_PAY error: %v", err)
		} else {
			log.Printf("[Callback] published PRODUCT_ORDER_PAY for order: %s", outTradeNo)
			return
		}
	}

	// Fallback: update order directly when MQ is unavailable
	if ctrl.db != nil {
		result := ctrl.db.Model(&shopdb.ProductOrderDO{}).
			Where("out_trade_no = ? AND state = ?", outTradeNo, string(enums.ORDER_NEW)).
			Update("state", string(enums.ORDER_PAY))
		if result.Error != nil {
			log.Printf("[Callback] direct order update error: %v", result.Error)
		} else {
			log.Printf("[Callback] order marked as PAY (direct): %s, rows=%d", outTradeNo, result.RowsAffected)
		}
	}
}

// xmlToMap parses an XML string into a map for signature verification.
func xmlToMap(xmlStr string) map[string]string {
	result := make(map[string]string)
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	var currentTag string
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			currentTag = t.Name.Local
		case xml.CharData:
			if currentTag != "" {
				result[currentTag] = string(t)
				currentTag = ""
			}
		}
	}
	return result
}
