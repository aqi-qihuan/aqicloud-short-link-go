package component

import "os"

// PayConfig holds credentials for Alipay and WeChat Pay.
// When sandbox fields are empty, strategies fall back to mock mode.
type PayConfig struct {
	// Alipay sandbox
	AliAppID      string
	AliPrivateKey string // PKCS8 PEM
	AliPublicKey  string // Alipay public key for signature verification
	AliGateway    string // sandbox: https://openapi-sandbox.dl.alipaydev.com/gateway.do
	AliNotifyURL  string

	// WeChat Pay sandbox
	WechatAppID     string
	WechatMchID     string
	WechatAPIKey    string // V2 API key for MD5 signing
	WechatNotifyURL string
}

func PayConfigFromEnv() PayConfig {
	return PayConfig{
		AliAppID:        getEnv("ALI_APP_ID", ""),
		AliPrivateKey:   getEnv("ALI_PRIVATE_KEY", ""),
		AliPublicKey:    getEnv("ALI_PUBLIC_KEY", ""),
		AliGateway:      getEnv("ALI_GATEWAY", "https://openapi-sandbox.dl.alipaydev.com/gateway.do"),
		AliNotifyURL:    getEnv("ALI_NOTIFY_URL", ""),
		WechatAppID:     getEnv("WECHAT_APP_ID", ""),
		WechatMchID:     getEnv("WECHAT_MCH_ID", ""),
		WechatAPIKey:    getEnv("WECHAT_API_KEY", ""),
		WechatNotifyURL: getEnv("WECHAT_NOTIFY_URL", ""),
	}
}

func (c *PayConfig) AliEnabled() bool {
	return c.AliAppID != "" && c.AliPrivateKey != "" && c.AliPublicKey != ""
}

func (c *PayConfig) WechatEnabled() bool {
	return c.WechatAppID != "" && c.WechatMchID != "" && c.WechatAPIKey != ""
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
