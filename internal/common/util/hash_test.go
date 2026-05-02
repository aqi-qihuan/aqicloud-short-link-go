package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMurmurHash3Guava(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint64
	}{
		{"空字符串", "", 0},
		{"ASCII字符串hello", "hello", 3619887497},
		{"短链码026m8O3a", "026m8O3a", 3431644148},
		{"长URL", "https://xdclass.net/#/coursedetail?video_id=1", 392181613},
		{"单字符", "a", 1867108634},
		{"数字字符串", "123456789", 2924794610},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MurmurHash3Guava(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMurmurHash3Guava_Deterministic(t *testing.T) {
	// 相同输入应产生相同输出
	s := "https://example.com/test"
	h1 := MurmurHash3Guava(s)
	h2 := MurmurHash3Guava(s)
	assert.Equal(t, h1, h2, "相同输入应产生相同哈希值")
}

func TestMurmurHash3Guava_DifferentInputs(t *testing.T) {
	// 不同输入应产生不同输出（碰撞概率极低）
	h1 := MurmurHash3Guava("https://example.com/a")
	h2 := MurmurHash3Guava("https://example.com/b")
	assert.NotEqual(t, h1, h2, "不同输入应产生不同哈希值")
}

func TestMurmurHash3Guava_ReturnType(t *testing.T) {
	// 验证返回值是 uint64 且高位为 0（padToLong 语义）
	h := MurmurHash3Guava("test")
	assert.Equal(t, uint64(0), h>>32, "高 32 位应为 0（padToLong 语义）")
}

func TestStringToUTF16LE(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
	}{
		{"空字符串", "", []byte{}},
		{"ASCII字母A", "A", []byte{0x41, 0x00}},
		{"ASCII字母AB", "AB", []byte{0x41, 0x00, 0x42, 0x00}},
		{"数字0", "0", []byte{0x30, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringToUTF16LE(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringToUTF16LE_ChineseChar(t *testing.T) {
	// 中文字符 '中' (U+4E2D) -> UTF-16 LE: 0x2D, 0x4E
	result := stringToUTF16LE("中")
	assert.Equal(t, []byte{0x2D, 0x4E}, result)
}

func TestStringToUTF16LE_NonBMP(t *testing.T) {
	// Emoji U+1F600 (😀) -> surrogate pair: D83D DE00
	result := stringToUTF16LE("😀")
	assert.Equal(t, 4, len(result), "non-BMP 字符应编码为 4 字节 surrogate pair")
	assert.Equal(t, byte(0x3D), result[0]) // high surrogate low byte
	assert.Equal(t, byte(0xD8), result[1]) // high surrogate high byte
}

func TestStringToUTF16LE_LengthIsDoubleUTF8ForASCII(t *testing.T) {
	// ASCII 字符串的 UTF-16 LE 长度应是字符数的 2 倍
	s := "hello world"
	result := stringToUTF16LE(s)
	assert.Equal(t, len(s)*2, len(result))
}

func TestJavaStringHashCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int32
	}{
		{"空字符串", "", 0},
		{"字母a", "a", 97},
		{"字母ab", "ab", 3105},
		{"字符串hello", "hello", 99162322},
		{"字符串test", "test", 3556498},
		{"数字字符串", "123456789", -1867378635},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JavaStringHashCode(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJavaStringHashCode_NonCommutative(t *testing.T) {
	// 验证 hashCode 与 Java 的行为一致：非交换律
	h1 := JavaStringHashCode("ab")
	h2 := JavaStringHashCode("ba")
	assert.NotEqual(t, h1, h2, "ab 和 ba 的 hashCode 应不同")
}

func TestAbsInt32Mod(t *testing.T) {
	tests := []struct {
		name     string
		hash     int32
		n        int
		expected int
	}{
		{"正数取模", 10, 3, 1},
		{"零取模", 0, 5, 0},
		{"负数取模", -7, 3, 0},
		{"MIN_VALUE取模", -2147483648, 3, 2},
		{"正数整除", 9, 3, 0},
		{"大质数取模", 123456789, 7, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AbsInt32Mod(tt.hash, tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAbsInt32Mod_AlwaysNonNegative(t *testing.T) {
	// AbsInt32Mod 应始终返回非负值
	testHashes := []int32{0, 1, -1, 2147483647, -2147483648, -100, 100}
	for _, h := range testHashes {
		result := AbsInt32Mod(h, 3)
		assert.GreaterOrEqual(t, result, 0,
			"AbsInt32Mod(%d, 3) = %d 应为非负", h, result)
	}
}
