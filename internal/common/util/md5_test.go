package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMD5_Empty(t *testing.T) {
	result := MD5("")
	assert.Equal(t, "D41D8CD98F00B204E9800998ECF8427E", result)
}

func TestMD5_Hello(t *testing.T) {
	result := MD5("hello")
	assert.Equal(t, "5D41402ABC4B2A76B9719D911017C592", result)
}

func TestMD5_Numeric(t *testing.T) {
	result := MD5("1234567890")
	assert.Equal(t, "E807F1FCF82D132F9BB018CA6738A19F", result)
}

func TestMD5_UppercaseHex(t *testing.T) {
	// 验证输出是大写十六进制（Java 兼容）
	result := MD5("test")
	for _, c := range result {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F'),
			"MD5 输出应为大写十六进制字符，实际: %c", c)
	}
}

func TestMD5_Deterministic(t *testing.T) {
	s := "consistent-input"
	assert.Equal(t, MD5(s), MD5(s), "相同输入应产生相同 MD5")
}

func TestMD5_DifferentInputs(t *testing.T) {
	assert.NotEqual(t, MD5("abc"), MD5("def"))
}

func TestRemoveUrlPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"带前缀", "123456789&https://example.com", "https://example.com"},
		{"无前缀", "https://example.com", "https://example.com"},
		{"空字符串", "", ""},
		{"多个&号", "123&abc&https://example.com", "abc&https://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveUrlPrefix(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddUrlPrefix(t *testing.T) {
	url := "https://example.com"
	result := AddUrlPrefix(url)
	assert.Contains(t, result, "&", "应包含 & 分隔符")
	assert.True(t, len(result) > len(url)+1, "长度应大于原 URL + 1")
}

func TestAddUrlPrefixVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"正常递增", "123456789&https://example.com", "123456790&https://example.com"},
		{"从1递增", "1&https://example.com", "2&https://example.com"},
		{"无前缀", "https://example.com", "https://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddUrlPrefixVersion(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
