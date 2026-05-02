package sms

import (
	"fmt"
	"log"
)

// Provider defines the interface for sending SMS messages.
type Provider interface {
	Send(phone string, templateCode string, params map[string]string) error
}

// LogProvider logs SMS messages instead of sending them (dev/test mode).
type LogProvider struct{}

func NewLogProvider() *LogProvider {
	return &LogProvider{}
}

func (p *LogProvider) Send(phone string, templateCode string, params map[string]string) error {
	log.Printf("[SMS-DEV] phone=%s template=%s params=%v", phone, templateCode, params)
	return nil
}

// AlibabaProvider sends SMS via Alibaba Cloud SMS service.
// Requires: ALIBABA_SMS_ACCESS_KEY, ALIBABA_SMS_ACCESS_SECRET, ALIBABA_SMS_SIGN_NAME
type AlibabaProvider struct {
	accessKey    string
	accessSecret string
	signName     string
}

func NewAlibabaProvider(accessKey, accessSecret, signName string) *AlibabaProvider {
	return &AlibabaProvider{
		accessKey:    accessKey,
		accessSecret: accessSecret,
		signName:     signName,
	}
}

func (p *AlibabaProvider) Send(phone string, templateCode string, params map[string]string) error {
	// TODO: integrate with Alibaba Cloud SMS SDK
	// API: dysmsapi.aliyuncs.com
	// Action: SendSms
	// Parameters: PhoneNumbers, SignName, TemplateCode, TemplateParam
	log.Printf("[SMS-Alibaba] phone=%s template=%s sign=%s params=%v", phone, templateCode, p.signName, params)
	return fmt.Errorf("Alibaba SMS not yet implemented - set SMS_PROVIDER=log for dev mode")
}

// TencentProvider sends SMS via Tencent Cloud SMS service.
type TencentProvider struct {
	appID     string
	secretID  string
	secretKey string
	signName  string
}

func NewTencentProvider(appID, secretID, secretKey, signName string) *TencentProvider {
	return &TencentProvider{
		appID:     appID,
		secretID:  secretID,
		secretKey: secretKey,
		signName:  signName,
	}
}

func (p *TencentProvider) Send(phone string, templateCode string, params map[string]string) error {
	// TODO: integrate with Tencent Cloud SMS SDK
	log.Printf("[SMS-Tencent] phone=%s template=%s sign=%s params=%v", phone, templateCode, p.signName, params)
	return fmt.Errorf("Tencent SMS not yet implemented - set SMS_PROVIDER=log for dev mode")
}

// NewProvider creates an SMS provider based on the provider name.
func NewProvider(provider string, config map[string]string) Provider {
	switch provider {
	case "alibaba":
		return NewAlibabaProvider(
			config["access_key"],
			config["access_secret"],
			config["sign_name"],
		)
	case "tencent":
		return NewTencentProvider(
			config["app_id"],
			config["secret_id"],
			config["secret_key"],
			config["sign_name"],
		)
	default:
		return NewLogProvider()
	}
}
