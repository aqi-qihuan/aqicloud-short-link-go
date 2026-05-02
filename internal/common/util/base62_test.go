package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeToBase62(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"零值", 0, "0"},
		{"1", 1, "1"},
		{"9", 9, "9"},
		{"10 -> a", 10, "a"},
		{"35 -> z", 35, "z"},
		{"36 -> A", 36, "A"},
		{"61 -> Z", 61, "Z"},
		{"62 -> 10", 62, "10"},
		{"100 -> 1C", 100, "1C"},
		{"3843 -> ZZ", 3843, "ZZ"},
		{"238327 -> ZZZ", 238327, "ZZZ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeToBase62(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEncodeToBase62_Charset(t *testing.T) {
	// 验证字符集为 0-9a-zA-Z
	for i := uint64(0); i < 62; i++ {
		result := EncodeToBase62(i)
		if i < 10 {
			// 0-9: 单字符
			assert.Len(t, result, 1)
			assert.Equal(t, byte('0'+i), result[0])
		} else if i < 36 {
			// a-z
			assert.Len(t, result, 1)
			assert.Equal(t, byte('a'+i-10), result[0])
		} else {
			// A-Z
			assert.Len(t, result, 1)
			assert.Equal(t, byte('A'+i-36), result[0])
		}
	}
}

func TestEncodeToBase62_OnlyValidChars(t *testing.T) {
	validChars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// 测试多个值确保只包含合法字符
	testValues := []uint64{0, 1, 61, 62, 100, 999999, 18446744073709551615}
	for _, v := range testValues {
		result := EncodeToBase62(v)
		for _, c := range result {
			assert.Contains(t, validChars, string(c),
				"字符 '%c' (0x%X) 不在 Base62 字符集中，输入值: %d", c, c, v)
		}
	}
}

func TestEncodeToBase62_MonotonicLength(t *testing.T) {
	// 数值越大，编码长度应该越长（或相等）
	prev := EncodeToBase62(0)
	for i := uint64(1); i < 1000; i++ {
		cur := EncodeToBase62(i)
		assert.GreaterOrEqual(t, len(cur), len(prev),
			"Base62 长度应该单调不减，i=%d", i)
		prev = cur
	}
}

func TestEncodeToBase62_MaxUint64(t *testing.T) {
	// max uint64 应产生最长编码且不 panic
	result := EncodeToBase62(18446744073709551615)
	assert.NotEmpty(t, result)
	assert.Greater(t, len(result), 10, "max uint64 编码长度应大于 10")
}

func TestEncodeToBase62_PowersOf62(t *testing.T) {
	// 62^1=62 -> "10", 62^2=3844 -> "100"
	assert.Equal(t, "10", EncodeToBase62(62))
	assert.Equal(t, "100", EncodeToBase62(3844))
	assert.Equal(t, "1000", EncodeToBase62(238328))
}
