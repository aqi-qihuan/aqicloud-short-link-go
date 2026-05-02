package util

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aqi/aqicloud-short-link-go/internal/common/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken_Success(t *testing.T) {
	user := &model.LoginUser{
		AccountNo: 123456789,
		Username:  "testuser",
		Mail:      "test@example.com",
		Phone:     "13800138000",
		HeadImg:   "https://example.com/avatar.jpg",
		Auth:      "admin",
	}

	token, err := GenerateToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, strings.HasPrefix(token, tokenPrefix), "token 应以 %s 开头", tokenPrefix)
}

func TestGenerateToken_ParseRoundtrip(t *testing.T) {
	user := &model.LoginUser{
		AccountNo: 658180183031197696, // 大整数
		Username:  "admin",
		Mail:      "admin@example.com",
		Phone:     "13900139000",
		HeadImg:   "https://example.com/img.png",
		Auth:      "superadmin",
	}

	token, err := GenerateToken(user)
	require.NoError(t, err)

	parsed, err := ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, user.AccountNo, parsed.AccountNo)
	assert.Equal(t, user.Username, parsed.Username)
	assert.Equal(t, user.Mail, parsed.Mail)
	assert.Equal(t, user.Phone, parsed.Phone)
	assert.Equal(t, user.HeadImg, parsed.HeadImg)
	assert.Equal(t, user.Auth, parsed.Auth)
}

func TestParseToken_InvalidToken(t *testing.T) {
	_, err := ParseToken("invalid.token.here")
	assert.Error(t, err, "无效 token 应返回错误")
}

func TestParseToken_EmptyToken(t *testing.T) {
	_, err := ParseToken("")
	assert.Error(t, err, "空 token 应返回错误")
}

func TestParseToken_WithPrefix(t *testing.T) {
	user := &model.LoginUser{AccountNo: 1, Username: "u"}
	token, err := GenerateToken(user)
	require.NoError(t, err)

	parsed, err := ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, int64(1), parsed.AccountNo)
}

func TestParseToken_WithoutPrefix(t *testing.T) {
	// 手动构造一个没有前缀的 JWT
	claims := jwt.MapClaims{
		"sub":        "xdclass",
		"account_no": float64(123),
		"username":   "rawuser",
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := jwtToken.SignedString([]byte(getJWTSecret()))
	require.NoError(t, err)

	// ParseToken 会 TrimPrefix，但如果没有前缀也能正常解析
	parsed, err := ParseToken(signed)
	require.NoError(t, err)
	assert.Equal(t, int64(123), parsed.AccountNo)
	assert.Equal(t, "rawuser", parsed.Username)
}

func TestParseToken_WrongSecret(t *testing.T) {
	// 用错误密钥签名
	claims := jwt.MapClaims{
		"sub":        "xdclass",
		"account_no": float64(123),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := jwtToken.SignedString([]byte("wrong-secret-key"))
	require.NoError(t, err)

	_, err = ParseToken(signed)
	assert.Error(t, err, "错误密钥签名的 token 应解析失败")
}

func TestParseToken_ExpiredToken(t *testing.T) {
	// 手动构造一个过期的 token
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":        "xdclass",
		"iat":        now.Add(-48 * time.Hour).Unix(),
		"exp":        now.Add(-24 * time.Hour).Unix(), // 已过期
		"account_no": float64(999),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := jwtToken.SignedString([]byte(getJWTSecret()))
	require.NoError(t, err)

	_, err = ParseToken(tokenPrefix + signed)
	assert.Error(t, err, "过期 token 应解析失败")
	assert.Contains(t, err.Error(), "expired", "错误应包含 expired 信息")
}

func TestParseToken_AccountNoPrecision(t *testing.T) {
	// 大整数 account_no (int64 最大值附近) 不应丢失精度
	user := &model.LoginUser{
		AccountNo: 9007199254740993, // 超出 float64 精度范围
		Username:  "precision_test",
	}

	token, err := GenerateToken(user)
	require.NoError(t, err)

	parsed, err := ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, int64(9007199254740993), parsed.AccountNo,
		"大整数 account_no 不应丢失精度")
}

func TestGenerateToken_EmptyUser(t *testing.T) {
	// 空用户也能生成 token（字段为空）
	user := &model.LoginUser{}
	token, err := GenerateToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	parsed, err := ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, int64(0), parsed.AccountNo)
	assert.Equal(t, "", parsed.Username)
}

func TestGenerateToken_EnvSecret(t *testing.T) {
	// 临时设置环境变量
	origSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", origSecret)

	os.Setenv("JWT_SECRET", "test-secret-from-env")

	user := &model.LoginUser{AccountNo: 1, Username: "envtest"}
	token, err := GenerateToken(user)
	require.NoError(t, err)

	// 用自定义密钥应能解析
	parsed, err := ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "envtest", parsed.Username)
}

func TestGetStringClaim(t *testing.T) {
	claims := jwt.MapClaims{
		"key1": "value1",
		"key2": 123, // 非字符串
		"key3": nil, // nil 值
	}

	assert.Equal(t, "value1", getStringClaim(claims, "key1"))
	assert.Equal(t, "", getStringClaim(claims, "key2"), "非字符串应返回空")
	assert.Equal(t, "", getStringClaim(claims, "key3"), "nil 应返回空")
	assert.Equal(t, "", getStringClaim(claims, "key4"), "不存在的 key 应返回空")
}
