package component

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aqi/aqicloud-short-link-go/internal/common/util"
)

// PayStrategy defines the interface for payment providers.
type PayStrategy interface {
	UnifiedOrder(payInfo *PayInfoVO) (string, error)
	Refund(payInfo *PayInfoVO) (string, error)
	QueryPayStatus(payInfo *PayInfoVO) (string, error)
	CloseOrder(payInfo *PayInfoVO) (string, error)
}

// PayInfoVO contains payment order information.
type PayInfoVO struct {
	OutTradeNo           string  `json:"outTradeNo"`
	PayFee               float64 `json:"payFee"` // yuan
	PayType              string  `json:"payType"`
	ClientType           string  `json:"clientType"` // APP/PC/H5
	Title                string  `json:"title"`
	Description          string  `json:"description"`
	OrderPayTimeoutMills int64   `json:"orderPayTimeoutMills"`
	AccountNo            int64   `json:"accountNo"`
}

// PayFactory resolves the correct PayStrategy by pay type.
type PayFactory struct {
	strategies map[string]PayStrategy
}

func NewPayFactory() *PayFactory {
	return &PayFactory{strategies: make(map[string]PayStrategy)}
}

func (f *PayFactory) Register(payType string, strategy PayStrategy) {
	f.strategies[payType] = strategy
}

func (f *PayFactory) GetStrategy(payType string) (PayStrategy, bool) {
	s, ok := f.strategies[payType]
	return s, ok
}

// ============================================================
// Alipay Sandbox Strategy
// ============================================================

type AliPayStrategy struct {
	cfg    *PayConfig
	client *http.Client
}

func NewAliPayStrategy(cfg *PayConfig) *AliPayStrategy {
	return &AliPayStrategy{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *AliPayStrategy) UnifiedOrder(payInfo *PayInfoVO) (string, error) {
	if !s.cfg.AliEnabled() {
		log.Printf("[AliPay] mock mode - no credentials configured")
		return fmt.Sprintf("https://openapi-sandbox.dl.alipaydev.com/mock/pay?out_trade_no=%s&total_amount=%.2f", payInfo.OutTradeNo, payInfo.PayFee), nil
	}

	// Build biz_content
	bizContent := map[string]interface{}{
		"out_trade_no":  payInfo.OutTradeNo,
		"total_amount":  fmt.Sprintf("%.2f", payInfo.PayFee),
		"subject":       payInfo.Title,
		"body":          payInfo.Description,
		"product_code":  "FAST_INSTANT_TRADE_PAY",
		"timeout_express": fmt.Sprintf("%dm", payInfo.OrderPayTimeoutMills/60000),
	}
	bizJSON, _ := json.Marshal(bizContent)

	// Build common params
	params := map[string]string{
		"app_id":      s.cfg.AliAppID,
		"method":      "alipay.trade.page.pay",
		"format":      "JSON",
		"return_url":  "",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  s.cfg.AliNotifyURL,
		"biz_content": string(bizJSON),
	}

	sign, err := signAlipayRSA2(params, s.cfg.AliPrivateKey)
	if err != nil {
		return "", fmt.Errorf("alipay sign: %w", err)
	}
	params["sign"] = sign

	// Build gateway URL
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	payURL := s.cfg.AliGateway + "?" + values.Encode()
	log.Printf("[AliPay] unified order: %s", payInfo.OutTradeNo)
	return payURL, nil
}

func (s *AliPayStrategy) QueryPayStatus(payInfo *PayInfoVO) (string, error) {
	if !s.cfg.AliEnabled() {
		return "TRADE_SUCCESS", nil // mock: always success
	}

	bizContent, _ := json.Marshal(map[string]string{
		"out_trade_no": payInfo.OutTradeNo,
	})

	params := map[string]string{
		"app_id":      s.cfg.AliAppID,
		"method":      "alipay.trade.query",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizContent),
	}

	sign, _ := signAlipayRSA2(params, s.cfg.AliPrivateKey)
	params["sign"] = sign

	resp, err := s.callAPI(params)
	if err != nil {
		return "", err
	}

	// Parse response to get trade_status
	var result struct {
		AlipayTradeQueryResponse struct {
			TradeStatus string `json:"trade_status"`
			Code        string `json:"code"`
			Msg         string `json:"msg"`
		} `json:"alipay_trade_query_response"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return "", fmt.Errorf("parse alipay query response: %w", err)
	}

	status := result.AlipayTradeQueryResponse.TradeStatus
	if status == "" {
		status = "TRADE_NOT_EXIST"
	}
	return status, nil
}

func (s *AliPayStrategy) CloseOrder(payInfo *PayInfoVO) (string, error) {
	if !s.cfg.AliEnabled() {
		return "SUCCESS", nil
	}

	bizContent, _ := json.Marshal(map[string]string{
		"out_trade_no": payInfo.OutTradeNo,
	})

	params := map[string]string{
		"app_id":      s.cfg.AliAppID,
		"method":      "alipay.trade.close",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizContent),
	}

	sign, _ := signAlipayRSA2(params, s.cfg.AliPrivateKey)
	params["sign"] = sign

	_, err := s.callAPI(params)
	return "SUCCESS", err
}

func (s *AliPayStrategy) Refund(payInfo *PayInfoVO) (string, error) {
	if !s.cfg.AliEnabled() {
		return "SUCCESS", nil
	}

	bizContent, _ := json.Marshal(map[string]interface{}{
		"out_trade_no":   payInfo.OutTradeNo,
		"refund_amount":  fmt.Sprintf("%.2f", payInfo.PayFee),
		"refund_reason":  "用户申请退款",
	})

	params := map[string]string{
		"app_id":      s.cfg.AliAppID,
		"method":      "alipay.trade.refund",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizContent),
	}

	sign, _ := signAlipayRSA2(params, s.cfg.AliPrivateKey)
	params["sign"] = sign

	_, err := s.callAPI(params)
	return "SUCCESS", err
}

func (s *AliPayStrategy) callAPI(params map[string]string) (string, error) {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	resp, err := s.client.PostForm(s.cfg.AliGateway, values)
	if err != nil {
		return "", fmt.Errorf("alipay request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read alipay response: %w", err)
	}
	return string(body), nil
}

// ============================================================
// WeChat Pay Sandbox Strategy (V2 API with MD5 signing)
// ============================================================

type WechatPayStrategy struct {
	cfg    *PayConfig
	client *http.Client
}

func NewWechatPayStrategy(cfg *PayConfig) *WechatPayStrategy {
	return &WechatPayStrategy{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// wechatV2Request is the XML request body for WeChat Pay V2.
type wechatV2Request struct {
	XMLName        xml.Name `xml:"xml"`
	AppID          string   `xml:"appid"`
	MchID          string   `xml:"mch_id"`
	NonceStr       string   `xml:"nonce_str"`
	Sign           string   `xml:"sign"`
	Body           string   `xml:"body"`
	OutTradeNo     string   `xml:"out_trade_no"`
	TotalFee       int      `xml:"total_fee"`
	SpbillCreateIP string   `xml:"spbill_create_ip"`
	NotifyURL      string   `xml:"notify_url"`
	TradeType      string   `xml:"trade_type"`
	SignType       string   `xml:"sign_type,omitempty"`
}

// wechatV2Response is the XML response from WeChat Pay V2.
type wechatV2Response struct {
	ReturnCode string `xml:"return_code"`
	ReturnMsg  string `xml:"return_msg"`
	ResultCode string `xml:"result_code"`
	PrepayID   string `xml:"prepay_id"`
	CodeURL    string `xml:"code_url"`
	Sign       string `xml:"sign"`
}

func (s *WechatPayStrategy) UnifiedOrder(payInfo *PayInfoVO) (string, error) {
	if !s.cfg.WechatEnabled() {
		log.Printf("[WeChat] mock mode - no credentials configured")
		return fmt.Sprintf("weixin://wxpay/bizpayurl?mock=1&out_trade_no=%s", payInfo.OutTradeNo), nil
	}

	totalFee := int(payInfo.PayFee * 100) // yuan -> fen

	params := map[string]string{
		"appid":            s.cfg.WechatAppID,
		"mch_id":           s.cfg.WechatMchID,
		"nonce_str":        util.GetStringNumRandom(32),
		"body":             payInfo.Title,
		"out_trade_no":     payInfo.OutTradeNo,
		"total_fee":        fmt.Sprintf("%d", totalFee),
		"spbill_create_ip": "127.0.0.1",
		"notify_url":       s.cfg.WechatNotifyURL,
		"trade_type":       "NATIVE",
		"sign_type":        "MD5",
	}
	params["sign"] = signWechatV2(params, s.cfg.WechatAPIKey)

	resp, err := s.callV2API("/pay/unifiedorder", params)
	if err != nil {
		return "", err
	}

	if resp.ReturnCode != "SUCCESS" {
		return "", fmt.Errorf("wechat unifiedorder: %s", resp.ReturnMsg)
	}
	if resp.ResultCode != "SUCCESS" {
		return "", fmt.Errorf("wechat unifiedorder result fail")
	}

	log.Printf("[WeChat] unified order: %s, code_url: %s", payInfo.OutTradeNo, resp.CodeURL)
	return resp.CodeURL, nil
}

func (s *WechatPayStrategy) QueryPayStatus(payInfo *PayInfoVO) (string, error) {
	if !s.cfg.WechatEnabled() {
		return "SUCCESS", nil // mock: always success
	}

	params := map[string]string{
		"appid":        s.cfg.WechatAppID,
		"mch_id":       s.cfg.WechatMchID,
		"out_trade_no": payInfo.OutTradeNo,
		"nonce_str":    util.GetStringNumRandom(32),
		"sign_type":    "MD5",
	}
	params["sign"] = signWechatV2(params, s.cfg.WechatAPIKey)

	resp, err := s.callV2API("/pay/orderquery", params)
	if err != nil {
		return "", err
	}

	if resp.ReturnCode != "SUCCESS" {
		return "", fmt.Errorf("wechat orderquery: %s", resp.ReturnMsg)
	}

	// Parse trade_state from XML response
	var fullResp struct {
		ReturnCode  string `xml:"return_code"`
		ResultCode  string `xml:"result_code"`
		TradeState  string `xml:"trade_state"`
		TradeStateDesc string `xml:"trade_state_desc"`
	}
	// Re-read response to get trade_state
	params2 := map[string]string{
		"appid":        s.cfg.WechatAppID,
		"mch_id":       s.cfg.WechatMchID,
		"out_trade_no": payInfo.OutTradeNo,
		"nonce_str":    util.GetStringNumRandom(32),
		"sign_type":    "MD5",
	}
	params2["sign"] = signWechatV2(params2, s.cfg.WechatAPIKey)
	xmlBody := buildXMLRequest(params2)
	httpResp, err := s.client.Post(sandboxURL("/pay/orderquery"), "application/xml", strings.NewReader(xmlBody))
	if err != nil {
		return "", err
	}
	defer httpResp.Body.Close()
	body, _ := io.ReadAll(httpResp.Body)
	if err := xml.Unmarshal(body, &fullResp); err != nil {
		return "", err
	}

	if fullResp.TradeState == "SUCCESS" {
		return "SUCCESS", nil
	}
	return fullResp.TradeState, nil
}

func (s *WechatPayStrategy) CloseOrder(payInfo *PayInfoVO) (string, error) {
	if !s.cfg.WechatEnabled() {
		return "SUCCESS", nil
	}

	params := map[string]string{
		"appid":        s.cfg.WechatAppID,
		"mch_id":       s.cfg.WechatMchID,
		"out_trade_no": payInfo.OutTradeNo,
		"nonce_str":    util.GetStringNumRandom(32),
		"sign_type":    "MD5",
	}
	params["sign"] = signWechatV2(params, s.cfg.WechatAPIKey)

	resp, err := s.callV2API("/pay/closeorder", params)
	if err != nil {
		return "", err
	}
	if resp.ReturnCode != "SUCCESS" {
		return "", fmt.Errorf("wechat closeorder: %s", resp.ReturnMsg)
	}
	return "SUCCESS", nil
}

func (s *WechatPayStrategy) Refund(payInfo *PayInfoVO) (string, error) {
	if !s.cfg.WechatEnabled() {
		return "SUCCESS", nil
	}

	totalFee := int(payInfo.PayFee * 100)
	params := map[string]string{
		"appid":         s.cfg.WechatAppID,
		"mch_id":        s.cfg.WechatMchID,
		"nonce_str":     util.GetStringNumRandom(32),
		"out_trade_no":  payInfo.OutTradeNo,
		"out_refund_no": "R" + payInfo.OutTradeNo,
		"total_fee":     fmt.Sprintf("%d", totalFee),
		"refund_fee":    fmt.Sprintf("%d", totalFee),
		"sign_type":     "MD5",
	}
	params["sign"] = signWechatV2(params, s.cfg.WechatAPIKey)

	resp, err := s.callV2API("/secapi/pay/refund", params)
	if err != nil {
		return "", err
	}
	if resp.ReturnCode != "SUCCESS" {
		return "", fmt.Errorf("wechat refund: %s", resp.ReturnMsg)
	}
	return "SUCCESS", nil
}

func (s *WechatPayStrategy) callV2API(path string, params map[string]string) (*wechatV2Response, error) {
	apiURL := sandboxURL(path)
	xmlBody := buildXMLRequest(params)

	httpResp, err := s.client.Post(apiURL, "application/xml", strings.NewReader(xmlBody))
	if err != nil {
		return nil, fmt.Errorf("wechat request: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read wechat response: %w", err)
	}

	var resp wechatV2Response
	if err := xml.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse wechat xml: %w", err)
	}
	return &resp, nil
}

// sandboxURL returns the WeChat Pay sandbox URL if sandbox mode, otherwise the production URL.
func sandboxURL(path string) string {
	// WeChat sandbox uses /sandboxnew/ prefix
	return "https://api.mch.weixin.qq.com/sandboxnew" + path
}

// buildXMLRequest builds an XML request body from params.
func buildXMLRequest(params map[string]string) string {
	var b strings.Builder
	b.WriteString("<xml>")
	for k, v := range params {
		b.WriteString("<" + k + ">" + v + "</" + k + ">")
	}
	b.WriteString("</xml>")
	return b.String()
}

// VerifyWechatV2Sign verifies a WeChat Pay V2 callback signature.
func VerifyWechatV2Sign(params map[string]string, apiKey string) bool {
	receivedSign := params["sign"]
	if receivedSign == "" {
		return false
	}
	delete(params, "sign")
	expected := signWechatV2(params, apiKey)
	return receivedSign == expected
}

// VerifyAlipaySign verifies an Alipay callback signature.
func VerifyAlipaySign(params map[string]string, publicKeyPEM string) bool {
	sign := params["sign"]
	if sign == "" {
		return false
	}
	ok, err := verifyAlipayRSA2(params, sign, publicKeyPEM)
	if err != nil {
		log.Printf("[AliPay] verify sign error: %v", err)
		return false
	}
	return ok
}
