package util

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/aqi/aqicloud-short-link-go/internal/common/model"
	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenPrefix  = "dcloud-link"
	tokenExpired = 7 * 24 * time.Hour
	subject      = "xdclass"
)

func getJWTSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "change-me-in-production"
}

// GenerateToken creates a JWT token compatible with the Java version.
func GenerateToken(loginUser *model.LoginUser) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":       subject,
		"iat":       now.Unix(),
		"exp":       now.Add(tokenExpired).Unix(),
		"head_img":  loginUser.HeadImg,
		"account_no": loginUser.AccountNo,
		"username":  loginUser.Username,
		"mail":      loginUser.Mail,
		"phone":     loginUser.Phone,
		"auth":      loginUser.Auth,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(getJWTSecret()))
	if err != nil {
		return "", err
	}
	return tokenPrefix + signed, nil
}

// ParseToken parses a JWT token (with "dcloud-link" prefix) and returns the LoginUser.
func ParseToken(tokenStr string) (*model.LoginUser, error) {
	tokenStr = strings.TrimPrefix(tokenStr, tokenPrefix)
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(getJWTSecret()), nil
	}, jwt.WithJSONNumber())
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	// account_no is a large int64; use json.Number to avoid float64 precision loss
	accountNo := int64(0)
	if v, ok := claims["account_no"]; ok && v != nil {
		switch n := v.(type) {
		case json.Number:
			if i, err := n.Int64(); err == nil {
				accountNo = i
			}
		case float64:
			accountNo = int64(n)
		case int64:
			accountNo = n
		}
	}
	return &model.LoginUser{
		AccountNo: accountNo,
		HeadImg:   getStringClaim(claims, "head_img"),
		Username:  getStringClaim(claims, "username"),
		Mail:      getStringClaim(claims, "mail"),
		Phone:     getStringClaim(claims, "phone"),
		Auth:      getStringClaim(claims, "auth"),
	}, nil
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	if v, ok := claims[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
