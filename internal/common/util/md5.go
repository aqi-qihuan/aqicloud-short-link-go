package util

import (
	"crypto/md5"
	"fmt"
)

// MD5 computes MD5 hash and returns uppercase hex string.
// Compatible with Java's CommonUtil.md5() which does toUpperCase().
func MD5(data string) string {
	h := md5.Sum([]byte(data))
	return fmt.Sprintf("%X", h) // %X = uppercase hex
}

// AddUrlPrefix prepends a snowflake ID to the URL: "snowflakeId&url"
func AddUrlPrefix(url string) string {
	return fmt.Sprintf("%d&%s", GenerateSnowflakeID(), url)
}

// RemoveUrlPrefix strips the snowflake ID prefix: "snowflakeId&url" -> "url"
func RemoveUrlPrefix(url string) string {
	idx := -1
	for i, c := range url {
		if c == '&' {
			idx = i
			break
		}
	}
	if idx >= 0 {
		return url[idx+1:]
	}
	return url
}

// AddUrlPrefixVersion increments the version prefix for collision retry.
// "123456789&https://..." -> "123456790&https://..."
func AddUrlPrefixVersion(url string) string {
	idx := -1
	for i, c := range url {
		if c == '&' {
			idx = i
			break
		}
	}
	if idx < 0 {
		return url
	}
	version := url[:idx]
	originalURL := url[idx+1:]
	// Parse version as int64 and increment
	var v int64
	for _, c := range version {
		v = v*10 + int64(c-'0')
	}
	v++
	return fmt.Sprintf("%d&%s", v, originalURL)
}
