package util

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetRandomCode_Length(t *testing.T) {
	tests := []int{0, 1, 6, 10, 100}
	for _, length := range tests {
		t.Run("长度"+string(rune('0'+length)), func(t *testing.T) {
			result := GetRandomCode(length)
			assert.Len(t, result, length)
		})
	}
}

func TestGetRandomCode_OnlyDigits(t *testing.T) {
	// 验证只包含数字字符（注意：故意匹配 Java bug，只有 0-8）
	result := GetRandomCode(1000)
	for _, c := range result {
		assert.True(t, c >= '0' && c <= '8',
			"GetRandomCode 应只产生 0-8 的字符，实际: %c", c)
	}
}

func TestGetRandomCode_NoDigitNine(t *testing.T) {
	// 文档说明：故意匹配 Java bug，nextInt(9) 排除了 '9'
	result := GetRandomCode(10000)
	assert.NotContains(t, result, "9", "不应包含 '9'（Java bug 兼容）")
}

func TestGetStringNumRandom_Length(t *testing.T) {
	tests := []int{0, 1, 10, 100}
	for _, length := range tests {
		result := GetStringNumRandom(length)
		assert.Len(t, result, length)
	}
}

func TestGetStringNumRandom_Charset(t *testing.T) {
	result := GetStringNumRandom(10000)
	validPattern := regexp.MustCompile(`^[0-9A-Za-z]+$`)
	assert.True(t, validPattern.MatchString(result),
		"应只包含 0-9A-Za-z，实际: %s", result[:50])
}

func TestGenerateUUID_Format(t *testing.T) {
	uuid := GenerateUUID()
	assert.Len(t, uuid, 32, "UUID 应为 32 字符（去横线）")
	assert.NotContains(t, uuid, "-", "UUID 不应包含横线")

	// 验证是十六进制字符
	validHex := regexp.MustCompile(`^[0-9a-f]{32}$`)
	assert.True(t, validHex.MatchString(uuid),
		"UUID 应为 32 位十六进制，实际: %s", uuid)
}

func TestGenerateUUID_Unique(t *testing.T) {
	seen := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		uuid := GenerateUUID()
		assert.False(t, seen[uuid], "UUID 重复: %s", uuid)
		seen[uuid] = true
	}
}

func TestGetCurrentTimestamp(t *testing.T) {
	before := time.Now().UnixMilli()
	ts := GetCurrentTimestamp()
	after := time.Now().UnixMilli()

	assert.GreaterOrEqual(t, ts, before, "时间戳应不早于调用前")
	assert.LessOrEqual(t, ts, after, "时间戳应不晚于调用后")
}

func TestGetCurrentTimestamp_MillisecondPrecision(t *testing.T) {
	ts := GetCurrentTimestamp()
	// 简单验证：时间戳 > 10^12（2001 年以来的毫秒时间戳）
	assert.Greater(t, ts, int64(1e12), "时间戳应为毫秒级精度")
}
