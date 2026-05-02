package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetRemainSecondsToday(t *testing.T) {
	remain := GetRemainSecondsToday()

	// 今天的剩余秒数应在 0 到 86399 之间
	assert.GreaterOrEqual(t, remain, 0, "剩余秒数不应为负")
	assert.LessOrEqual(t, remain, 86399, "剩余秒数不应超过 86399")

	// 粗略验证：当前时间与 endOfDay 的差应在 1 秒误差内
	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	expected := int(endOfDay.Sub(now).Seconds())
	assert.InDelta(t, expected, remain, 1, "误差应在 1 秒内")
}

func TestGetStartOfDay(t *testing.T) {
	start := GetStartOfDay()
	now := time.Now()

	assert.Equal(t, now.Year(), start.Year())
	assert.Equal(t, now.Month(), start.Month())
	assert.Equal(t, now.Day(), start.Day())
	assert.Equal(t, 0, start.Hour())
	assert.Equal(t, 0, start.Minute())
	assert.Equal(t, 0, start.Second())
	assert.Equal(t, 0, start.Nanosecond())
}

func TestIsToday(t *testing.T) {
	assert.True(t, IsToday(time.Now()), "当前时间应是今天")
	assert.True(t, IsToday(time.Now().Add(-1*time.Hour)), "1小时前应是今天")
}

func TestIsToday_Yesterday(t *testing.T) {
	yesterday := time.Now().Add(-24 * time.Hour)
	assert.False(t, IsToday(yesterday), "昨天不应是今天")
}

func TestIsToday_Tomorrow(t *testing.T) {
	tomorrow := time.Now().Add(24 * time.Hour)
	assert.False(t, IsToday(tomorrow), "明天不应是今天")
}

func TestIsToday_SameDayDifferentTime(t *testing.T) {
	now := time.Now()
	// 同一天不同时间
	morning := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	night := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	assert.True(t, IsToday(morning), "今天 00:00:00 应是今天")
	assert.True(t, IsToday(night), "今天 23:59:59 应是今天")
}

func TestGetStartOfDay_IsMidnight(t *testing.T) {
	start := GetStartOfDay()
	// start 应该是午夜
	assert.Equal(t, 0, start.Hour())
	assert.Equal(t, 0, start.Minute())
	assert.Equal(t, 0, start.Second())
}
