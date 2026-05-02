package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aqi/aqicloud-short-link-go/internal/common/constant"
	"github.com/aqi/aqicloud-short-link-go/internal/common/sms"
	"github.com/aqi/aqicloud-short-link-go/internal/common/util"
	"github.com/redis/go-redis/v9"
)

const (
	SendCodeTypeRegister = "USER_REGISTER"
	CaptchaKeyPrefix     = "account-service:captcha:"
	CaptchaTTL           = 10 * time.Minute
	CodeTTL              = 10 * time.Minute
	CodeRateLimit        = 60 * time.Second
)

type NotifyService struct {
	rdb      *redis.Client
	smsProv  sms.Provider
}

func NewNotifyService(rdb *redis.Client, smsProv sms.Provider) *NotifyService {
	return &NotifyService{rdb: rdb, smsProv: smsProv}
}

// GenerateCaptcha creates a 4-digit numeric captcha code.
func (s *NotifyService) GenerateCaptcha() string {
	return util.GetRandomCode(4)
}

// SaveCaptcha stores the captcha in Redis keyed by MD5(ip+userAgent).
func (s *NotifyService) SaveCaptcha(key, captcha string) error {
	return s.rdb.Set(context.Background(), CaptchaKeyPrefix+key, captcha, CaptchaTTL).Err()
}

// ValidateCaptcha checks the captcha against Redis cache and deletes it on success.
func (s *NotifyService) ValidateCaptcha(key, captcha string) bool {
	redisKey := CaptchaKeyPrefix + key
	val, err := s.rdb.Get(context.Background(), redisKey).Result()
	if err != nil {
		return false
	}
	if val != captcha {
		return false
	}
	s.rdb.Del(context.Background(), redisKey)
	return true
}

// SendCode generates and stores a 6-digit SMS verification code.
func (s *NotifyService) SendCode(sendCodeType, to string) error {
	redisKey := constant.FormatCheckCodeKey(sendCodeType, to)

	// Rate limit check
	lastSent, err := s.rdb.Get(context.Background(), redisKey+":ts").Int64()
	if err == nil && time.Now().UnixMilli()-lastSent < CodeRateLimit.Milliseconds() {
		return fmt.Errorf("please wait %d seconds before requesting again", int(CodeRateLimit.Seconds()))
	}

	code := util.GetRandomCode(6)
	value := fmt.Sprintf("%s_%d", code, time.Now().UnixMilli())

	pipe := s.rdb.Pipeline()
	pipe.Set(context.Background(), redisKey, value, CodeTTL)
	pipe.Set(context.Background(), redisKey+":ts", time.Now().UnixMilli(), CodeTTL)
	_, err = pipe.Exec(context.Background())
	if err != nil {
		return fmt.Errorf("save code to redis failed: %w", err)
	}

	// Send SMS via provider
	templateCode := "SMS_REGISTER_CODE"
	params := map[string]string{"code": code}
	if err := s.smsProv.Send(to, templateCode, params); err != nil {
		return fmt.Errorf("send SMS failed: %w", err)
	}
	return nil
}

// CheckCode validates the SMS verification code.
func (s *NotifyService) CheckCode(sendCodeType, phone, code string) bool {
	redisKey := constant.FormatCheckCodeKey(sendCodeType, phone)
	val, err := s.rdb.Get(context.Background(), redisKey).Result()
	if err != nil {
		return false
	}
	// Format: "code_timestamp"
	if len(val) < 7 { // min: "123456_"
		return false
	}
	storedCode := val[:6]
	if storedCode != code {
		return false
	}
	s.rdb.Del(context.Background(), redisKey)
	return true
}
