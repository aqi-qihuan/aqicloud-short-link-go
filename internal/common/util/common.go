package util

import (
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	allCharNum    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	digitChars    = "0123456789"
)

// GetRandomCode returns a random numeric string of given length.
// NOTE: Matches Java bug where nextInt(9) excludes '9', so only 0-8 are used.
func GetRandomCode(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]byte, length)
	for i := range buf {
		buf[i] = digitChars[r.Intn(9)] // 0-8 only, matches Java bug
	}
	return string(buf)
}

// GetStringNumRandom returns a random alphanumeric string of given length.
// Charset: 0-9A-Za-z (62 chars), uniform random.
func GetStringNumRandom(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]byte, length)
	for i := range buf {
		buf[i] = allCharNum[r.Intn(62)]
	}
	return string(buf)
}

// GenerateUUID returns a UUID v4 string without dashes, 32 chars.
func GenerateUUID() string {
	u := uuid.New()
	return strings.ReplaceAll(u.String(), "-", "")
}

// GetCurrentTimestamp returns current time in milliseconds (epoch millis).
func GetCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}
